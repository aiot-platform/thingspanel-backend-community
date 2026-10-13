// Package devicelog 设备统一消息日志（device_log 表）。
//
// 每条设备上行报文（遥测/属性/事件/响应）、平台下行报文（遥测控制/属性设置/属性获取/命令）
// 以及设备上下线都追加写入一行，用于按设备检索完整的消息流水。
//
// 设计要点：
//   - Record 非阻塞：写入缓冲队列，队列满时丢弃并计数，绝不阻塞 MQTT/业务处理链路；
//   - 后台协程批量写库（按条数或时间间隔刷新）；
//   - 表为 TimescaleDB hypertable（按天分块），保留期由 retention policy + 定时 DELETE 双重保证；
//   - 本包只依赖 gorm/logrus，供 uplink、downlink、dal 等包调用而不产生循环依赖。
package devicelog

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TableName 日志表名
const TableName = "device_log"

// 方向
const (
	DirectionUp   = "up"   // 设备 → 平台
	DirectionDown = "down" // 平台 → 设备
)

// 消息类型
const (
	TypeTelemetry            = "telemetry"              // 遥测上报
	TypeAttribute            = "attribute"              // 属性上报
	TypeEvent                = "event"                  // 事件上报
	TypeStatus               = "status"                 // 上下线
	TypeTelemetryControl     = "telemetry_control"      // 遥测控制下发
	TypeAttributeSet         = "attribute_set"          // 属性设置下发
	TypeAttributeGet         = "attribute_get"          // 属性获取下发
	TypeCommand              = "command"                // 命令下发
	TypeCommandResponse      = "command_response"       // 命令响应
	TypeAttributeSetResponse = "attribute_set_response" // 属性设置响应
)

// 状态：上下线（TypeStatus）不填；其余类型均为 ok / failed
const (
	StatusOK     = "ok"
	StatusFailed = "failed"
)

// 单条报文最大保存字节数，超出部分截断，避免异常大包撑爆日志表
const maxPayloadBytes = 64 * 1024

// Entry 一条设备日志
type Entry struct {
	TS        time.Time // 为空时取 Record 调用时间
	DeviceID  string    // 必填
	TenantID  string    // 为空时写库前按 device_id 自动补齐
	MsgType   string
	Direction string
	MessageID string
	Status    string // 上下线为空，其余为 ok / failed
	Error     string
	Payload   []byte // 原始报文；非 JSON 会被包装成 {"_raw": "..."}
}

// Config 配置
type Config struct {
	Enabled       bool
	RetentionDays int
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
}

// DefaultConfig 默认配置：保留 3 天
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		RetentionDays: 3,
		BufferSize:    10000,
		BatchSize:     500,
		FlushInterval: time.Second,
	}
}

// row 与 device_log 表列对应
type row struct {
	TS        time.Time `gorm:"column:ts"`
	DeviceID  string    `gorm:"column:device_id"`
	TenantID  string    `gorm:"column:tenant_id"`
	MsgType   string    `gorm:"column:msg_type"`
	Direction string    `gorm:"column:direction"`
	MessageID *string   `gorm:"column:message_id"`
	Status    *string   `gorm:"column:status"`
	Payload   *string   `gorm:"column:payload"`
	Error     *string   `gorm:"column:error"`
}

// Writer 异步批量写入器
type Writer struct {
	db     *gorm.DB
	logger *logrus.Logger
	cfg    Config

	ch      chan Entry
	stopCh  chan struct{}
	doneCh  chan struct{}
	dropped atomic.Int64

	// 仅由写入协程访问，无需加锁
	tenantCache map[string]string
}

var current atomic.Pointer[Writer]

// New 创建写入器（未启动）
func New(db *gorm.DB, logger *logrus.Logger, cfg Config) *Writer {
	def := DefaultConfig()
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = def.RetentionDays
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = def.BufferSize
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = def.BatchSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = def.FlushInterval
	}
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &Writer{
		db:          db,
		logger:      logger,
		cfg:         cfg,
		ch:          make(chan Entry, cfg.BufferSize),
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		tenantCache: make(map[string]string),
	}
}

// Record 记录一条设备日志（非阻塞；写入器未启用时直接忽略）
func Record(e Entry) {
	w := current.Load()
	if w == nil || e.DeviceID == "" {
		return
	}
	if e.TS.IsZero() {
		e.TS = time.Now()
	}
	// 复制报文，避免调用方复用底层数组
	if len(e.Payload) > 0 {
		n := len(e.Payload)
		if n > maxPayloadBytes+1 {
			n = maxPayloadBytes + 1 // 多留 1 字节用于判断是否截断
		}
		e.Payload = append([]byte(nil), e.Payload[:n]...)
	}
	select {
	case w.ch <- e:
	default:
		w.dropped.Add(1)
	}
}

// EnsureSchema 建表、转 hypertable、建索引、设置保留策略（幂等）
func (w *Writer) EnsureSchema() error {
	createTable := `CREATE TABLE IF NOT EXISTS ` + TableName + ` (
	ts          TIMESTAMPTZ  NOT NULL,
	device_id   VARCHAR(36)  NOT NULL,
	tenant_id   VARCHAR(36)  NOT NULL DEFAULT '',
	msg_type    VARCHAR(32)  NOT NULL,
	direction   VARCHAR(8)   NOT NULL,
	message_id  VARCHAR(64),
	status      VARCHAR(16),
	payload     JSONB,
	error       TEXT
)`
	if err := w.db.Exec(createTable).Error; err != nil {
		return fmt.Errorf("create table %s: %w", TableName, err)
	}
	_ = w.db.Exec(`COMMENT ON TABLE ` + TableName + ` IS '设备统一消息日志：上下行报文与上下线，按保留期自动清理'`).Error

	// 兼容早期版本建的表：identifier 列已废弃（method 仍保留在 payload 中）
	if err := w.db.Exec(`ALTER TABLE ` + TableName + ` DROP COLUMN IF EXISTS identifier`).Error; err != nil {
		w.logger.WithError(err).Warn("device_log: drop column identifier failed")
	}

	// 转为 hypertable（按天分块）。没有 timescaledb 扩展时降级为普通表，靠定时 DELETE 清理
	if err := w.db.Exec(`SELECT create_hypertable('` + TableName + `', 'ts', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE)`).Error; err != nil {
		w.logger.WithError(err).Warn("device_log: create_hypertable failed, fallback to plain table")
	}

	if err := w.db.Exec(`CREATE INDEX IF NOT EXISTS idx_device_log_device_ts ON ` + TableName + ` (device_id, ts DESC)`).Error; err != nil {
		return fmt.Errorf("create index on %s: %w", TableName, err)
	}

	// 保留策略：先移除旧策略再按当前配置添加，修改 retention_days 后重启即可生效
	_ = w.db.Exec(`SELECT remove_retention_policy('` + TableName + `', if_exists => TRUE)`).Error
	if err := w.db.Exec(fmt.Sprintf(`SELECT add_retention_policy('%s', INTERVAL '%d days')`, TableName, w.cfg.RetentionDays)).Error; err != nil {
		w.logger.WithError(err).Warn("device_log: add_retention_policy failed, rely on periodic DELETE")
	}

	w.logger.Infof("device_log: schema ready, retention=%d days", w.cfg.RetentionDays)
	return nil
}

// Start 启动后台写入协程，并设为全局写入器
func (w *Writer) Start() {
	current.Store(w)
	go w.run()
}

// Stop 停止写入：不再接收新日志，刷完队列后退出
func (w *Writer) Stop(timeout time.Duration) {
	current.CompareAndSwap(w, nil)
	close(w.stopCh)
	select {
	case <-w.doneCh:
	case <-time.After(timeout):
		w.logger.Warn("device_log: stop timeout, some logs may be lost")
	}
}

func (w *Writer) run() {
	defer close(w.doneCh)

	flushTicker := time.NewTicker(w.cfg.FlushInterval)
	defer flushTicker.Stop()
	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()
	dropTicker := time.NewTicker(time.Minute)
	defer dropTicker.Stop()

	batch := make([]Entry, 0, w.cfg.BatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w.write(batch)
		batch = batch[:0]
	}

	w.cleanup()

	for {
		select {
		case e := <-w.ch:
			batch = append(batch, e)
			if len(batch) >= w.cfg.BatchSize {
				flush()
			}
		case <-flushTicker.C:
			flush()
		case <-cleanupTicker.C:
			w.cleanup()
		case <-dropTicker.C:
			if n := w.dropped.Swap(0); n > 0 {
				w.logger.Warnf("device_log: buffer full, dropped %d entries in last minute", n)
			}
		case <-w.stopCh:
			for {
				select {
				case e := <-w.ch:
					batch = append(batch, e)
					if len(batch) >= w.cfg.BatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

// write 批量写库
func (w *Writer) write(entries []Entry) {
	w.resolveTenants(entries)

	rows := make([]row, 0, len(entries))
	for i := range entries {
		e := &entries[i]
		rows = append(rows, row{
			TS:        e.TS,
			DeviceID:  e.DeviceID,
			TenantID:  e.TenantID,
			MsgType:   e.MsgType,
			Direction: e.Direction,
			MessageID: nullable(e.MessageID, 64),
			Status:    nullable(e.Status, 16),
			Payload:   normalizePayload(e.Payload),
			Error:     nullable(e.Error, 0),
		})
	}

	if err := w.db.Table(TableName).CreateInBatches(rows, 500).Error; err != nil {
		// 批量失败（如个别报文含 JSONB 不接受的 \u0000）时逐条重试，只丢弃坏行
		w.logger.WithError(err).Warnf("device_log: batch insert %d entries failed, retry one by one", len(rows))
		failed := 0
		for i := range rows {
			if err := w.db.Table(TableName).Create(&rows[i]).Error; err != nil {
				failed++
			}
		}
		if failed > 0 {
			w.logger.Errorf("device_log: %d/%d entries dropped after retry", failed, len(rows))
		}
	}
}

// resolveTenants 为缺少 tenant_id 的日志按 device_id 批量补齐
func (w *Writer) resolveTenants(entries []Entry) {
	var missing []string
	seen := make(map[string]struct{})
	for i := range entries {
		e := &entries[i]
		if e.TenantID != "" {
			w.tenantCache[e.DeviceID] = e.TenantID
			continue
		}
		if t, ok := w.tenantCache[e.DeviceID]; ok {
			e.TenantID = t
			continue
		}
		if _, ok := seen[e.DeviceID]; !ok {
			seen[e.DeviceID] = struct{}{}
			missing = append(missing, e.DeviceID)
		}
	}

	if len(missing) > 0 {
		var devices []struct {
			ID       string
			TenantID string
		}
		if err := w.db.Table("devices").Select("id, tenant_id").Where("id IN ?", missing).Scan(&devices).Error; err != nil {
			w.logger.WithError(err).Warn("device_log: resolve tenant failed")
		}
		for _, d := range devices {
			w.tenantCache[d.ID] = d.TenantID
		}
		for i := range entries {
			if entries[i].TenantID == "" {
				entries[i].TenantID = w.tenantCache[entries[i].DeviceID]
			}
		}
	}

	// 防止缓存无限增长
	if len(w.tenantCache) > 50000 {
		w.tenantCache = make(map[string]string)
	}
}

// cleanup 删除超出保留期的数据（与 retention policy 互为兜底，精确到小时）
func (w *Writer) cleanup() {
	res := w.db.Exec(`DELETE FROM `+TableName+` WHERE ts < now() - make_interval(days => ?)`, w.cfg.RetentionDays)
	if res.Error != nil {
		w.logger.WithError(res.Error).Warn("device_log: cleanup failed")
		return
	}
	if res.RowsAffected > 0 {
		w.logger.Infof("device_log: cleaned %d expired entries", res.RowsAffected)
	}
}

// normalizePayload 报文转为可写入 JSONB 的文本
func normalizePayload(p []byte) *string {
	if len(p) == 0 {
		return nil
	}
	var s string
	switch {
	case len(p) > maxPayloadBytes:
		b, _ := json.Marshal(map[string]interface{}{
			"_truncated": true,
			"_raw":       string(p[:maxPayloadBytes]),
		})
		s = string(b)
	case json.Valid(p):
		s = string(p)
	default:
		b, _ := json.Marshal(map[string]interface{}{"_raw": string(p)})
		s = string(b)
	}
	return &s
}

// nullable 空串写 NULL；max>0 时按字符截断
func nullable(s string, max int) *string {
	if s == "" {
		return nil
	}
	if max > 0 {
		if r := []rune(s); len(r) > max {
			s = string(r[:max])
		}
	}
	return &s
}

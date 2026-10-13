package app

import (
	"time"

	"project/internal/devicelog"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// DeviceLogServiceWrapper 设备统一日志服务
type DeviceLogServiceWrapper struct {
	writer *devicelog.Writer
}

// Name 返回服务名称
func (s *DeviceLogServiceWrapper) Name() string {
	return "device_log 设备日志服务"
}

// Start 启动写入协程
func (s *DeviceLogServiceWrapper) Start() error {
	if s.writer != nil {
		s.writer.Start()
	}
	return nil
}

// Stop 刷完缓冲后停止
func (s *DeviceLogServiceWrapper) Stop() error {
	if s.writer != nil {
		s.writer.Stop(10 * time.Second)
	}
	return nil
}

// WithDeviceLogService 注册设备日志服务。
// 需放在 WithDatabase 之后、WithStorageService/WithFlowService 之前：
// 这样它先于业务链路启动，并在关停时最后停止，保证剩余日志能刷入数据库。
func WithDeviceLogService() Option {
	return func(a *Application) error {
		cfg := devicelog.DefaultConfig()
		if viper.IsSet("device_log.enabled") {
			cfg.Enabled = viper.GetBool("device_log.enabled")
		}
		if viper.IsSet("device_log.retention_days") {
			cfg.RetentionDays = viper.GetInt("device_log.retention_days")
		}
		if viper.IsSet("device_log.buffer_size") {
			cfg.BufferSize = viper.GetInt("device_log.buffer_size")
		}
		if viper.IsSet("device_log.batch_size") {
			cfg.BatchSize = viper.GetInt("device_log.batch_size")
		}
		if viper.IsSet("device_log.flush_interval") {
			cfg.FlushInterval = time.Duration(viper.GetInt("device_log.flush_interval")) * time.Millisecond
		}

		wrapper := &DeviceLogServiceWrapper{}
		if !cfg.Enabled {
			logrus.Info("device_log is disabled in config")
			a.RegisterService(wrapper)
			return nil
		}

		writer := devicelog.New(a.DB, a.Logger, cfg)
		// 建表失败不阻断平台启动，只是不记录设备日志
		if err := writer.EnsureSchema(); err != nil {
			logrus.WithError(err).Error("device_log: init schema failed, device log disabled")
			a.RegisterService(wrapper)
			return nil
		}

		wrapper.writer = writer
		a.RegisterService(wrapper)

		ensureDeviceLogMenu(a.DB)
		return nil
	}
}

// ensureDeviceLogMenu 幂等地在「系统管理」下注册「设备日志查询」菜单（仅系统管理员可见）。
// 已存在（含管理员在菜单管理里改过名称/排序）时不覆盖。
func ensureDeviceLogMenu(db *gorm.DB) {
	err := db.Exec(`
INSERT INTO sys_ui_elements
  (id, parent_id, element_code, element_type, orders, param1, param2, param3,
   authority, description, created_at, remark, multilingual, route_path)
SELECT '7d3a9c52-1e4b-4f6a-9b0d-6c2e8f1a5d47', p.id, 'management_device-log', 3, 45,
       '/management/device-log', 'mdi:file-document-outline', 'self',
       '["SYS_ADMIN"]'::json, '设备日志查询', now(), '', 'route.management_device-log', 'view.management_device-log'
FROM sys_ui_elements p
WHERE p.element_code = 'management' AND p.parent_id = '0'
  AND NOT EXISTS (SELECT 1 FROM sys_ui_elements WHERE element_code = 'management_device-log')
LIMIT 1`).Error
	if err != nil {
		logrus.WithError(err).Warn("device_log: ensure menu failed")
	}
}

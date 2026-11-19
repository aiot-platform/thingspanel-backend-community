// Package solution 提供「行业方案」的数据库 schema 级隔离能力。
//
// 每个行业方案在同一个 PostgreSQL 实例内拥有独立 schema（如 sol_pet_medical），
// 方案业务表存放在各自 schema 中。本包按 schema 名惰性打开并缓存独立的 *gorm.DB，
// 连接串追加 search_path=<schema>,public：
//   - 裸表名优先命中方案 schema 内的表；
//   - 未命中时回退 public（例如 sys_version 等平台表）。
//
// 不使用「单连接 SET search_path」的方案，因为连接池会复用连接、search_path 会被污染。
package solution

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// schemaNameRe 校验方案 schema 名，避免 SQL 注入（schema 名会直接拼进连接串）。
var schemaNameRe = regexp.MustCompile(`^sol_[a-z0-9_]+$`)

var (
	mu    sync.Mutex
	cache = make(map[string]*gorm.DB)
)

// DB 返回指定方案 schema 的 *gorm.DB（惰性创建并缓存）。
func DB(schemaName string) (*gorm.DB, error) {
	if !schemaNameRe.MatchString(schemaName) {
		return nil, fmt.Errorf("invalid solution schema name: %q", schemaName)
	}

	mu.Lock()
	defer mu.Unlock()

	if db, ok := cache[schemaName]; ok {
		return db, nil
	}

	db, err := open(schemaName)
	if err != nil {
		return nil, err
	}
	cache[schemaName] = db
	return db, nil
}

// MustDB 与 DB 相同，但忽略错误（仅用于 schema 一定存在的场景）。
func MustDB(schemaName string) *gorm.DB {
	db, err := DB(schemaName)
	if err != nil {
		logrus.Errorf("获取方案 schema 连接失败 (%s): %v", schemaName, err)
	}
	return db
}

// CloseAll 关闭所有已缓存的方案连接，进程退出时调用。
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()
	for name, db := range cache {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		delete(cache, name)
	}
}

func open(schemaName string) (*gorm.DB, error) {
	host := viper.GetString("db.psql.host")
	if host == "" {
		host = "localhost"
	}
	port := viper.GetInt("db.psql.port")
	if port == 0 {
		port = 5432
	}
	dbName := viper.GetString("db.psql.dbname")
	user := viper.GetString("db.psql.username")
	password := viper.GetString("db.psql.password")
	timeZone := viper.GetString("db.psql.time_zone")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}
	if dbName == "" || user == "" {
		return nil, fmt.Errorf("database configuration is incomplete for solution schema %s", schemaName)
	}

	// search_path 追加 <schema>,public：裸表名优先命中方案表，其次回退平台表。
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable TimeZone=%s search_path=%s,public",
		host, port, dbName, user, password, timeZone, schemaName,
	)

	sqlLogger := logger.Default.LogMode(logger.Warn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 sqlLogger,
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接方案 schema %s 失败: %w", schemaName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取方案 schema %s 原生连接失败: %w", schemaName, err)
	}
	// 每个方案独立连接池，上限保持较低，避免方案数量增长后连接数失控。
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logrus.Infof("方案 schema 连接已建立: %s", schemaName)
	return db, nil
}

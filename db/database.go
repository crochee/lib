package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/crochee/lirity"
	"github.com/crochee/lirity/logger"
)

var (
	NotFound           = gorm.ErrRecordNotFound
	ErrNotRowsAffected = errors.New("0 rows affected")
	ErrDuplicate       = "1062: Duplicate"
)

type Option struct {
	Debug bool

	MaxOpenConn int
	MaxIdleConn int

	User     string
	Password string
	IP       string
	Port     string
	Database string
	Charset  string

	Timeout         time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ConnMaxLifetime time.Duration
}

// New with context.Context returns DB
func New(ctx context.Context, opts ...func(*Option)) (*DB, error) {
	var o Option
	for _, opt := range opts {
		opt(&o)
	}
	client, err := gorm.Open(mysql.Open(Dsn(&o)),
		&gorm.Config{
			PrepareStmt: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 不考虑表名单复数变化
			},
			DisableForeignKeyConstraintWhenMigrating: true,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	if err != nil {
		return nil, err
	}
	session := &gorm.Session{Context: ctx}
	if o.Debug { // 是否显示sql语句
		session.Logger = client.Logger.LogMode(glogger.Info)
	}
	client = client.Session(session)

	var sqlDB *sql.DB
	if sqlDB, err = client.DB(); err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB.SetMaxOpenConns(o.MaxOpenConn)        // 默认值0，无限制
	sqlDB.SetMaxIdleConns(o.MaxIdleConn)        // 默认值2
	sqlDB.SetConnMaxLifetime(o.ConnMaxLifetime) // 默认值0，永不过期
	return &DB{DB: client, debug: o.Debug}, nil
}

type DB struct {
	*gorm.DB
	debug bool
}

type SessionOption struct {
	SlowThreshold time.Duration
	Colorful      bool
	LevelFunc     func(glogger.LogLevel, bool) glogger.LogLevel
}

// With options to set orm logger
func (d *DB) With(ctx context.Context, opts ...func(*SessionOption)) *DB {
	log := logger.From(ctx)
	o := &SessionOption{
		SlowThreshold: 10 * time.Second,
		Colorful:      false,
		LevelFunc:     getLevel,
	}
	for _, opt := range opts {
		opt(o)
	}
	return &DB{DB: d.Session(&gorm.Session{
		Context: ctx,
		Logger: NewLog(log, glogger.Config{
			SlowThreshold: o.SlowThreshold,
			Colorful:      o.Colorful,
			LogLevel:      o.LevelFunc(glogger.Warn, d.debug),
		}),
	}),
		debug: d.debug,
	}
}

// Close impl io.Closer to defer close db pool
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	lirity.Close(sqlDB)
	return nil
}

func getLevel(l glogger.LogLevel, debug bool) glogger.LogLevel {
	if debug {
		return glogger.Info
	}
	return l
}

func Dsn(opt *Option) string {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		opt.User, opt.Password, opt.IP, opt.Port, opt.Database, opt.Charset, true, "UTC")
	if opt.Timeout != 0 {
		uri += fmt.Sprintf("&timeout=%s", opt.Timeout)
	}
	if opt.ReadTimeout != 0 {
		uri += fmt.Sprintf("&readTimeout=%s", opt.ReadTimeout)
	}
	if opt.ReadTimeout != 0 {
		uri += fmt.Sprintf("&writeTimeout=%s", opt.ReadTimeout)
	}
	return uri
}

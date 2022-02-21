package db

import (
	"context"
	"runtime"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crochee/lirity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var c *DB

// Init init database
func Init(ctx context.Context, opts ...func(*Option)) (err error) {
	c, err = New(ctx, opts...)
	runtime.SetFinalizer(c, ClientClose)
	return
}

// With call db
func With(ctx context.Context, opts ...func(*SessionOption)) *DB {
	return c.With(ctx, opts...)
}

func ClientClose(db *DB) {
	if db == nil {
		return
	}
	lirity.Close(db)
}

// Mock new a mock 解除测试对数据库等中间件的依赖
func Mock() (sqlmock.Sqlmock, error) {
	// 创建sqlmock
	slqDb, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	// 结合gorm、sqlmock
	var client *gorm.DB
	if client, err = gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      slqDb,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 不考虑表名单复数变化
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}); err != nil {
		return nil, err
	}
	c = &DB{DB: client, debug: true}
	return mock, err
}

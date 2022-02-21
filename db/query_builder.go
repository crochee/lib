package db

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/crochee/lirity/variable"
)

type builderOption struct {
}

type BuilderOption func(*builderOption)

// SQLBuilder 将参数组装成 gorm.DB 即预处理的sql语句
type SQLBuilder interface {
	Build(ctx context.Context, query *gorm.DB, opts ...BuilderOption) *gorm.DB
}

// Sort 排序
type Sort struct {
	// 给多个字段排序 created_at, id asc => order by created_at desc, id asc
	SortField string `form:"sort" json:"sort" binding:"omitempty,order"`
}

func (s *Sort) Build(_ context.Context, query *gorm.DB, _ ...BuilderOption) *gorm.DB {
	// SortField 给多个字段排序
	// created_at, id asc => order by created_at desc, id asc
	if s.SortField != "" {
		sorts := strings.Split(s.SortField, ",")
		for _, field := range sorts {
			// 如果排序没有明确要按asc或desc来排序，则按照默认排序(倒序)
			if !strings.HasSuffix(field, "asc") && !strings.HasSuffix(field, "desc") {
				query = query.Order(fmt.Sprintf("%s desc", field))
				continue
			}
			query = query.Order(field)
		}
	}

	return query
}

// Pagination 分页
type Pagination struct {
	Page   int   `form:"page" json:"page" binding:"omitempty,min=0"`
	Size   int   `form:"size" json:"size" binding:"omitempty,min=-1"`
	Offset int   `json:"-" form:"offset"`
	Limit  int   `json:"-" form:"limit"`
	Total  int64 `json:"total"`
}

func (p *Pagination) Build(_ context.Context, query *gorm.DB, _ ...BuilderOption) *gorm.DB {
	query.Count(&p.Total)
	if p.Limit == 0 {
		if p.Size == 0 {
			p.Size = variable.DefaultPageSize
		}
		p.Limit = p.Size
	}
	if p.Offset == 0 {
		if p.Page == 0 {
			p.Page = variable.DefaultPageNum
		}
		p.Offset = (p.Page - 1) * p.Limit
	}
	return query.Limit(p.Limit).Offset(p.Offset)
}

type Primary struct {
	Sort
	Pagination
}

func (p *Primary) Build(ctx context.Context, query *gorm.DB, opts ...BuilderOption) *gorm.DB {
	query = p.Sort.Build(ctx, query, opts...)
	return p.Pagination.Build(ctx, query, opts...)
}

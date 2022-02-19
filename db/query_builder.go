package db

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/crochee/lirity/variable"
)

type sqlBuilderOption struct {
}

type SqlBuilderOption func(*sqlBuilderOption)

// SqlBuilder 将参数组装成 gorm.DB 即预处理的sql语句
type SqlBuilder interface {
	Build(ctx context.Context, query *gorm.DB, opts ...SqlBuilderOption) *gorm.DB
}

// Sort 排序
type Sort struct {
	// 给多个字段排序 created_at, id asc => order by created_at desc, id asc
	SortField string `form:"sort" json:"sort" binding:"omitempty,order"`
}

func (s *Sort) Build(_ context.Context, query *gorm.DB, _ ...SqlBuilderOption) *gorm.DB {
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
	PageNum  int   `form:"page_num" json:"page_num" binding:"omitempty,min=0"`
	PageSize int   `form:"page_size" json:"page_size" binding:"omitempty,min=-1"`
	Total    int64 `json:"total"`
}

func (p *Pagination) Build(_ context.Context, query *gorm.DB, _ ...SqlBuilderOption) *gorm.DB {
	query.Count(&p.Total)
	// -1表示全量查询
	if p.PageSize == -1 {
		return query
	}
	if p.PageNum == 0 {
		p.PageNum = variable.DefaultPageNum
	}
	if p.PageSize == 0 {
		p.PageSize = variable.DefaultPageSize
	}
	return query.Limit(p.PageSize).Offset((p.PageNum - 1) * p.PageSize)
}

type Primary struct {
	Sort
	Pagination
}

func (p *Primary) Build(ctx context.Context, query *gorm.DB, opts ...SqlBuilderOption) *gorm.DB {
	query = p.Sort.Build(ctx, query, opts...)
	return p.Pagination.Build(ctx, query, opts...)
}

package query

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// Filter 过滤条件
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Sort 排序条件
type Sort struct {
	Field string
	Desc  bool
}

// Pagination 分页条件
type Pagination struct {
	Page     int
	PageSize int
}

// Builder 查询构建器
type Builder struct {
	db *gorm.DB
}

// NewBuilder 创建查询构建器
func NewBuilder(db *gorm.DB) *Builder {
	return &Builder{db: db}
}

// ApplyFilters 应用过滤条件
func (b *Builder) ApplyFilters(filters []Filter) *Builder {
	for _, filter := range filters {
		switch filter.Operator {
		case "=":
			b.db = b.db.Where(filter.Field+" = ?", filter.Value)
		case "!=":
			b.db = b.db.Where(filter.Field+" != ?", filter.Value)
		case ">":
			b.db = b.db.Where(filter.Field+" > ?", filter.Value)
		case ">=":
			b.db = b.db.Where(filter.Field+" >= ?", filter.Value)
		case "<":
			b.db = b.db.Where(filter.Field+" < ?", filter.Value)
		case "<=":
			b.db = b.db.Where(filter.Field+" <= ?", filter.Value)
		case "LIKE":
			b.db = b.db.Where(filter.Field+" LIKE ?", "%"+filter.Value.(string)+"%")
		case "IN":
			b.db = b.db.Where(filter.Field+" IN ?", filter.Value)
		}
	}
	return b
}

// ApplySorts 应用排序条件
func (b *Builder) ApplySorts(sorts []Sort) *Builder {
	for _, sort := range sorts {
		if sort.Desc {
			b.db = b.db.Order(sort.Field + " DESC")
		} else {
			b.db = b.db.Order(sort.Field + " ASC")
		}
	}
	return b
}

// ApplySort 应用排序条件（别名）
func (b *Builder) ApplySort(sorts []Sort) *Builder {
	return b.ApplySorts(sorts)
}

// Where 添加where条件
func (b *Builder) Where(query interface{}, args ...interface{}) *Builder {
	b.db = b.db.Where(query, args...)
	return b
}

// ApplyPagination 应用分页条件
func (b *Builder) ApplyPagination(pagination *Pagination) *Builder {
	if pagination != nil && pagination.Page > 0 && pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		b.db = b.db.Offset(offset).Limit(pagination.PageSize)
	}
	return b
}

// Count 获取总数
func (b *Builder) Count(count *int64) *Builder {
	b.db.Count(count)
	return b
}

// Find 查询结果
func (b *Builder) Find(dest interface{}) error {
	return b.db.Find(dest).Error
}

// DB 获取底层DB
func (b *Builder) DB() *gorm.DB {
	return b.db
}

// Options 查询选项
type Options struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// NewOptionsFromQuery 从查询参数创建选项
func NewOptionsFromQuery(c *gin.Context) *Options {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	return &Options{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
}

// ApplyOptions 应用查询选项到GORM查询
func ApplyOptions(db *gorm.DB, opts *Options) *gorm.DB {
	if opts == nil {
		return db
	}
	
	// 应用排序
	if opts.SortBy != "" {
		order := opts.SortBy
		if opts.SortOrder == "asc" {
			order += " ASC"
		} else {
			order += " DESC"
		}
		db = db.Order(order)
	}
	
	// 应用分页
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		db = db.Offset(offset).Limit(opts.PageSize)
	}
	
	return db
}

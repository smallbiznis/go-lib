package pagination

import "gorm.io/gorm"

type Pagination struct {
	PerPage int    `form:"per_page" validate:"min=1,max=250"`
	Page    int    `form:"page" validate:"min=1"`
	SortBy  string `form:"sort_by"`
	OrderBy string `form:"order_by"`
}

func (p Pagination) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p.PerPage > 0 {
			db.Limit(p.PerPage)
		}

		offset := (p.Page - 1) * p.PerPage
		return db.Offset(offset)
	}
}

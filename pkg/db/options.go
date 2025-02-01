package db

import (
	"time"

	"gorm.io/gorm"
)

type Options func(*gorm.DB) *gorm.DB

func WithOrderByASC() Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order("id ASC")
	}
}

func WithOrderByDesc() Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order("id DESC")
	}
}

func WithOffset(offset int) Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset(offset)
	}
}

func WithCreatedBefore(t time.Time) Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("gmt_create < ?", t)
	}
}

func WithLimit(limit int) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if limit == 0 {
			// `LIMIT 0` statement will return 0 rows, it's meaningless.
			return tx
		}
		return tx.Limit(limit)
	}
}

func WithIDIn(ids ...int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		// e.g. `WHERE id IN (1, 2, 3)`
		return tx.Where("id IN ?", ids)
	}
}

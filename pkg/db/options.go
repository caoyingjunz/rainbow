package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Options func(*gorm.DB) *gorm.DB

func WithTagOrderByDESC() Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order("tag DESC")
	}
}

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

func WithModifyOrderByDesc() Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order("gmt_modified DESC")
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

func WithCreatedAfter(t time.Time) Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("gmt_create > ?", t)
	}
}

func WithPublic() Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("is_public = 1")
	}
}

func WithEnable(enable int) Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("enable = ?", enable)
	}
}

func WithFailTimes(t int) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if t == 0 {
			return tx
		}
		return tx.Where("fail_times <= ?", t)
	}
}

func WithRole(role int) Options {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where("role = ?", role)
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

func WithIDStrIn(ids ...string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		// e.g. `WHERE id IN (1, 2, 3)`
		return tx.Where("id IN ?", ids)
	}
}

func WithErrorTask(onlyErr bool) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if !onlyErr {
			return tx
		}
		return tx.Where("status = ?", "Error")
	}
}

func WithName(name string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return tx
		}
		return tx.Where("name = ?", name)
	}
}

func WithRainbowdName(name string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return tx
		}
		return tx.Where("rainbowd_name = ?", name)
	}
}

func WithPath(path string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(path) == 0 {
			return tx
		}
		return tx.Where("path = ?", path)
	}
}

func WithNameIn(names ...string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(names) == 0 {
			return tx
		}

		// e.g. `WHERE id IN (1, 2, 3)`
		return tx.Where("name IN ?", names)
	}
}

func WithLabelIn(labels ...string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(labels) == 0 {
			return tx
		}
		return tx.Where("label IN ?", labels)
	}
}

func WithId(id int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if id == 0 {
			return tx
		}
		return tx.Where("id = ?", id)
	}
}

func WithUser(userId string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(userId) == 0 {
			return tx
		}
		return tx.Where("user_id = ?", userId)
	}
}

func WithArch(arch string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(arch) == 0 {
			return tx
		}
		return tx.Where("arch = ?", arch)
	}
}

func WithTask(taskId int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if taskId == 0 {
			return tx
		}
		return tx.Where("task_id = ?", taskId)
	}
}

func WithBuild(buildId int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if buildId == 0 {
			return tx
		}
		return tx.Where("build_id = ?", buildId)
	}
}

func WithSubscribe(subId int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if subId == 1 { // 默认获取全部, 否则按 subscribe_id 获取列表
			return tx
		}
		return tx.Where("subscribe_id = ?", subId)
	}
}

func WithTaskLike(taskId int64) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if taskId == 0 {
			return tx
		}
		return tx.Where("task_ids like ?", "%"+fmt.Sprintf("%d", taskId)+"%")
	}
}

func WithNameLike(name string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return tx
		}
		return tx.Where("name like ?", "%"+name+"%")
	}
}

func WithPathLike(path string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(path) == 0 {
			return tx
		}
		return tx.Where("path like ?", "%"+path+"%")
	}
}

func WithTagLike(tag string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(tag) == 0 {
			return tx
		}
		return tx.Where("tag like ?", "%"+tag+"%")
	}
}

func WithNamespace(ns string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(ns) == 0 {
			return tx
		}
		return tx.Where("namespace = ?", ns)
	}
}

func WithAgent(agent string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(agent) == 0 {
			return tx
		}
		return tx.Where("agent_name = ?", agent)
	}
}

func WithRef(ref string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if ref == "0" || ref == "1" {
			return tx.Where("owner_ref = ?", ref)
		}
		return tx
	}
}

func WithStatus(status string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return tx
		}
		return tx.Where("status = ?", status)
	}
}

func WithType(Type string) Options {
	return func(tx *gorm.DB) *gorm.DB {
		if len(Type) == 0 {
			return tx
		}
		return tx.Where("Type = ?", Type)
	}
}

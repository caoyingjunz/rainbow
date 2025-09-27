package timeutil

import (
	"fmt"
	"time"
)

func ToTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	hours := diff.Hours()
	days := hours / 24

	switch {
	case days >= 365:
		years := int(days / 365)
		return fmt.Sprintf("%d 年前", years)
	case days >= 30:
		months := int(days / 30)
		return fmt.Sprintf("%d 个月前", months)
	case days >= 1:
		return fmt.Sprintf("%d 天前", int(days))
	case hours >= 1:
		return fmt.Sprintf("%d 小时前", int(hours))
	case diff.Minutes() >= 1:
		return fmt.Sprintf("%d 分钟前", int(diff.Minutes()))
	default:
		return "刚刚"
	}
}

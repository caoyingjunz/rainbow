package util

import "strings"

func InSlice(s string, ss []string) bool {
	for _, sl := range ss {
		if sl == s {
			return true
		}
	}

	return false
}

func TrimAndFilter(input []string) []string {
	result := make([]string, 0, len(input))
	seen := make(map[string]struct{})

	for _, s := range input {
		trimmed := strings.TrimSpace(s) // 去除前后空格
		if trimmed == "" {              // 过滤空字符串或纯空格
			continue
		}
		if _, exists := seen[trimmed]; !exists { // 检查是否已存在
			result = append(result, trimmed)
			seen[trimmed] = struct{}{} // 标记为已处理
		}
	}

	return result
}

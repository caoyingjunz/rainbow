package util

func InSlice(s string, ss []string) bool {
	for _, sl := range ss {
		if sl == s {
			return true
		}
	}

	return false
}

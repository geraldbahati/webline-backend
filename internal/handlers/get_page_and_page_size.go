package handlers

import "strconv"

func GetPageAndPageSize(pageStr string, pageSizeStr string) (int32, int32, error) {
	var page int32 = 0
	var pageSize int32 = 0

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = int32(p)
	}

	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = int32(ps)
	}

	return page, pageSize, nil
}

package service

const (
	defaultPageSize int32 = 20
	maxPageSize     int32 = 100
)

// clampPagination normalises page size and page number to safe values.
func clampPagination(pageSize, page int32) (int32, int32) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page <= 0 {
		page = 1
	}
	return pageSize, page
}

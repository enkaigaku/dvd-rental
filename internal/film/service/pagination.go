package service

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func clampPagination(pageSize, page int32) (int32, int32) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page <= 0 {
		page = 1
	}
	return pageSize, page
}

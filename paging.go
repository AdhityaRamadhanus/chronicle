package chronicle

//PagingOptions is a struct used as pagination option to get entities
type PagingOptions struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

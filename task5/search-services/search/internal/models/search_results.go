package models

type SearchResults struct {
	Result []SearchResultUnit
	Total  int
}

type SearchResultUnit struct {
	ID  int
	URL string
}

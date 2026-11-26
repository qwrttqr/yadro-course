package models

type SearchResults struct {
	Result []SearchResultUnit
}

type SearchResultUnit struct {
	ID  int
	URL string
}

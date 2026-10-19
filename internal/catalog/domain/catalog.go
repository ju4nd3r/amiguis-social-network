package domain

import "errors"

var (
	ErrCatalogItemNotFound = errors.New("catalog item not found")
	ErrInvalidRating       = errors.New("rating must be between 0 and 5")
)

type CatalogCategory string

const (
	CategoryMovie  CatalogCategory = "movie"
	CategorySeries CatalogCategory = "series"
	CategoryBook   CatalogCategory = "book"
)

type CatalogItem struct {
	ID            string          `json:"id"`
	Category      CatalogCategory `json:"category"`
	Title         string          `json:"title"`
	OriginalTitle string          `json:"original_title,omitempty"`
	Year          int             `json:"year"`
	Description   string          `json:"description"`
	PosterURL     string          `json:"poster_url"`
	AverageRating float64         `json:"average_rating"`
	RatingsCount  int             `json:"ratings_count"`
	Cast          []string        `json:"cast"`
	Producer      string          `json:"producer"`
	Creator       string          `json:"creator"`
	Genres        []string        `json:"genres"`
	Duration      string          `json:"duration"`
	Country       string          `json:"country"`
}

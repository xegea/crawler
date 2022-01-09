package scraper

type Movie struct {
	Title         map[string]string
	Url           string
	ContentRating string
	Type          string
	Description   map[string]string
	Genre         string
	Image         string
	ReleaseDate   int64
	Director      []string
	Actors        []string
	Trailer       []Trailer
}

type Trailer struct {
	Name         map[string]string
	Description  map[string]string
	Url          string
	ThumbnailUrl string
}

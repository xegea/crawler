package scraper

type Movie struct {
	Title         string
	Url           string
	ContentRating string
	Type          string
	Description   string
	Genre         string
	Image         string
	ReleaseDate   int64
	Director      []string
	Actors        []string
	Trailer       []Trailer
}

type Trailer struct {
	Name         string
	Description  string
	Url          string
	ThumbnailUrl string
}

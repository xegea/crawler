package streaming

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
	Updated       int64
}

type Trailer struct {
	Name         string
	Description  string
	Url          string
	ThumbnailUrl string
}

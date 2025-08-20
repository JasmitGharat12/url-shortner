package models

type Request struct {
	URL string `json:"url"` // URL to be shortened
}

type Response struct {
	ShortURL string `json:"short_url"` // Generated short URL
}

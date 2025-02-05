package dto

type OriginalURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ShortedURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

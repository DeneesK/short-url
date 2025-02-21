package dto

type OriginalURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type LongUrl struct {
	LongURL   string `json:"long_url"`
	IsDeleted bool   `json:"is_deleted"`
}

type ShortedURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

type URL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type UpdateTask struct {
	UserID string `json:"user_id"`
	ID     string `json:"id"`
}

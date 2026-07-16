package review

type Review struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	AuthorID    int    `json:"author_id"`
	TargetID    int    `json:"target_id"`
	Note        int    `json:"note"` // 1-5
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

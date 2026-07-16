package exchange

type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"` // pending, accepted, rejected, cancelled, completed
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

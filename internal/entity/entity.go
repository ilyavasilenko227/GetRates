package entity

type Order struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}

type Depth struct {
	Timestamp int64 `json:"timestamp"`
	Asks      Order `json:"asks"`
	Bids      Order `json:"bids"`
}

type DepthRequest struct {
	Timestamp int64   `json:"timestamp"`
	Asks      []Order `json:"asks"`
	Bids      []Order `json:"bids"`
}

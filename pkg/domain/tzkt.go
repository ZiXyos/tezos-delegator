package domain

type TzktApiDelegationsResponse struct {
	Type                string   `json:"type"`
	ID                  int64    `json:"id"`
	Level               int64    `json:"level"`
	Timestamp           string   `json:"timestamp"`
	Block               string   `json:"block"`
	Hash                string   `json:"hash"`
	Counter             int64    `json:"counter"`
	Initiator           *Account `json:"initiator,omitempty"`
	Sender              *Account `json:"sender,omitempty"`
	SenderCodeHash      *int     `json:"senderCodeHash,omitempty"`
	Nonce               *int     `json:"nonce,omitempty"`
	GasLimit            int      `json:"gasLimit"`
	GasUsed             int      `json:"gasUsed"`
	StorageLimit        int      `json:"storageLimit"`
	BakerFee            int64    `json:"bakerFee"`
	Amount              int64    `json:"amount"`
	StakingUpdatesCount *int     `json:"stakingUpdatesCount,omitempty"`
	PrevDelegate        *Account `json:"prevDelegate,omitempty"`
	NewDelegate         *Account `json:"newDelegate,omitempty"`
	Status              string   `json:"status"`
	Errors              []Error  `json:"errors,omitempty"`
	Quote               *Quote   `json:"quote,omitempty"`
}

type Account struct {
	Alias   string `json:"alias,omitempty"`
	Address string `json:"address"`
}

type Error struct {
	Type string `json:"type"`
}

type Quote struct {
	BTC float64 `json:"btc"`
	EUR float64 `json:"eur"`
	USD float64 `json:"usd"`
	CNY float64 `json:"cny"`
	JPY float64 `json:"jpy"`
	KRW float64 `json:"krw"`
	ETH float64 `json:"eth"`
	GBP float64 `json:"gbp"`
}

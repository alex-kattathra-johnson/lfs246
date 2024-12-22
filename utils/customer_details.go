package utils

import "encoding/json"

type CustomerDetails struct {
	Id                  string `json:"id"`
	CustomerName        string `json:"custName"`
	WalletAmount        int    `json:"walletAmount"`
	WalletAmountBlocked int    `json:"walletAmountBlocked"`
}

func (cd CustomerDetails) String() string {
	b, _ := json.Marshal(cd)
	return string(b)
}

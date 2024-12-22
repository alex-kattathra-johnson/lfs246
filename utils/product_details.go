package utils

import "encoding/json"

type ProductDetails struct {
	Id               string `json:"id"`
	ProductName      string `json:"prodName"`
	ProductAvailable int    `json:"productAvailable"`
	ProductBlocked   int    `json:"productBlocked"`
}

func (pd ProductDetails) String() string {
	b, _ := json.Marshal(pd)
	return string(b)
}

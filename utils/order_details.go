package utils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin/binding"
)

type OrderStatus int

const (
	ORDERSTATUS_NEW OrderStatus = iota
	ORDERSTATUS_IN_PROGRESS
	ORDERSTATUS_CUSTOMER_CONFIRMED
	ORDERSTATUS_PRODUCT_CONFIRMED
	ORDERSTATUS_CONFIRMED
)

func (os OrderStatus) String() string {
	switch os {
	case ORDERSTATUS_NEW:
		return "NEW"
	case ORDERSTATUS_IN_PROGRESS:
		return "IN_PROGRESS"
	case ORDERSTATUS_CUSTOMER_CONFIRMED:
		return "CUSTOMER_CONFIRMED"
	case ORDERSTATUS_PRODUCT_CONFIRMED:
		return "PRODUCT_CONFIRMED"
	case ORDERSTATUS_CONFIRMED:
		return "CONFIRMED"
	}
	return ""
}

type OrderDetails struct {
	Id                  uint        `json:"id"`
	CustomerId          uint        `json:"custId"`
	ProductId           uint        `json:"prodId"`
	Amount              uint        `json:"amount"`
	ProductCount        uint        `json:"prodCount"`
	ProductOrderStatus  OrderStatus `json:"product_order_status"`
	CustomerOrderStatus OrderStatus `json:"customer_order_status"`
	OrderStatus         OrderStatus `json:"order_status"`
}

func (od OrderDetails) CallCustomer() {
	data, _ := json.Marshal(od)
	http.Post(
		"http://customer-service.default.svc.cluster.local/customers/blockAmount",
		binding.MIMEJSON,
		bytes.NewBuffer(data),
	)
}

func (od OrderDetails) CallProduct() {
	data, _ := json.Marshal(od)
	http.Post(
		"http://product-service.default.svc.cluster.local/products/block",
		binding.MIMEJSON,
		bytes.NewBuffer(data),
	)
}

func (od OrderDetails) ConfirmOrder() {
	data, _ := json.Marshal(od)
	http.Post(
		"http://order-service.default.svc.cluster.local/orders/confirm",
		binding.MIMEJSON,
		bytes.NewBuffer(data),
	)
}

type OrderResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
}

func (od OrderDetails) String() string {
	b, _ := json.Marshal(od)
	return string(b)
}

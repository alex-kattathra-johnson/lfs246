package utils

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
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
	Id                  string      `json:"id"`
	CustomerId          string      `json:"custId"`
	ProductId           string      `json:"prodId"`
	Amount              int         `json:"amount"`
	ProductCount        int         `json:"prodCount"`
	ProductOrderStatus  OrderStatus `json:"product_order_status"`
	CustomerOrderStatus OrderStatus `json:"customer_order_status"`
	OrderStatus         OrderStatus `json:"order_status"`
}

func (od OrderDetails) SendTo(topic string) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	sender, err := kafka_sarama.NewSender([]string{"my-cluster-kafka-bootstrap.kafka:9092"}, saramaConfig, topic)
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer sender.Close(context.Background())

	c, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(od.OrderStatus.String())
	e.SetSource("http://localhost")
	if err = e.SetData(cloudevents.ApplicationJSON, od); err != nil {
		log.Fatalf("failed to set data: %s", err)
	}

	if result := c.Send(kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())), e); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}

type OrderResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
}

func (od OrderDetails) String() string {
	b, _ := json.Marshal(od)
	return string(b)
}

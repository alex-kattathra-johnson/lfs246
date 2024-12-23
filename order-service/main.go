package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/alex-kattathra-johnson/lfs246/utils"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

var orderDetailsRepo *badger.DB

func init() {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatalf("could not create database: %s", err)
	}
	orderDetailsRepo = db
}

func main() {
	r := gin.Default()
	r.POST("/orders/place", placeOrder)

	r.POST("/orders/confirm", confirmOrder)

	r.Run()
}

func placeOrder(c *gin.Context) {
	var od utils.OrderDetails
	if err := c.Bind(&od); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("order-service :: Order Request Received from revision :: %s", os.Getenv("VERSION"))
	log.Infof("order-service :: Order Request Received :: %s", od)
	log.Infof("order-service :: Order Status in Request :: %s", od.OrderStatus)

	if od.OrderStatus == utils.ORDERSTATUS_NEW {
		if err := orderDetailsRepo.Update(func(txn *badger.Txn) error {
			data, err := json.Marshal(od)
			if err != nil {
				return err
			}
			return txn.Set([]byte(od.Id), data)
		}); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		od.SendTo("customer")
		od.SendTo("product")

		c.JSON(http.StatusAccepted, utils.OrderResponse{
			Status:        "SUCCESS",
			StatusMessage: "Request Processed Successfully",
		})
		return
	}

	c.JSON(http.StatusAccepted, utils.OrderResponse{
		Status:        utils.ORDERSTATUS_IN_PROGRESS.String(),
		StatusMessage: "Request Processed Successfully",
	})
}

func confirmOrder(c *gin.Context) {
	od := new(utils.OrderDetails)
	if err := c.Bind(od); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("order-service :: Final Order Confirmation :: %s", od)
	log.Infof("order-service :: Order Status :: %s", od.OrderStatus)

	if err := orderDetailsRepo.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(od.Id))
		if err != nil {
			return err
		}

		var data []byte
		if err := item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		}); err != nil {
			return err
		}

		var order utils.OrderDetails
		if err := json.Unmarshal(data, &order); err != nil {
			return err
		}

		switch od.OrderStatus {
		case utils.ORDERSTATUS_CUSTOMER_CONFIRMED:
			order.OrderStatus = utils.ORDERSTATUS_IN_PROGRESS
			order.CustomerOrderStatus = od.OrderStatus
		case utils.ORDERSTATUS_PRODUCT_CONFIRMED:
			order.OrderStatus = utils.ORDERSTATUS_IN_PROGRESS
			order.ProductOrderStatus = od.OrderStatus
		}

		if order.CustomerOrderStatus == utils.ORDERSTATUS_CUSTOMER_CONFIRMED && order.ProductOrderStatus == utils.ORDERSTATUS_PRODUCT_CONFIRMED {
			order.OrderStatus = utils.ORDERSTATUS_CONFIRMED
			order.SendTo("customer")
			order.SendTo("product")
		}

		data, err = json.Marshal(order)
		if err != nil {
			return err
		}

		return txn.Set([]byte(order.Id), data)
	}); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", binding.MIMEPlain)
	c.String(http.StatusOK, "200 OK.")
}

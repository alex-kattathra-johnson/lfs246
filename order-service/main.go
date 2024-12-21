package main

import (
	"net/http"
	"sync"

	"github.com/alex-kattathra-johnson/lfs246/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

var orderDetailsRepo sync.Map

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

	log.Infof("order-service :: Order Request Received :: %s", od)
	log.Infof("order-service :: Order Status in Request :: %s", od.OrderStatus)

	if od.OrderStatus == utils.ORDERSTATUS_NEW {
		orderDetailsRepo.Store(od.Id, od)

		od.CallCustomer()
		od.CallProduct()

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

	data, ok := orderDetailsRepo.Load(od.Id)

	if ok {
		order := data.(utils.OrderDetails)

		switch od.OrderStatus {
		case utils.ORDERSTATUS_CUSTOMER_CONFIRMED:
			order.OrderStatus = utils.ORDERSTATUS_IN_PROGRESS
			order.CustomerOrderStatus = od.OrderStatus
		case utils.ORDERSTATUS_PRODUCT_CONFIRMED:
			order.OrderStatus = utils.ORDERSTATUS_IN_PROGRESS
			order.ProductOrderStatus = od.OrderStatus
		}

		orderDetailsRepo.Store(od.Id, order)

		if order.CustomerOrderStatus == utils.ORDERSTATUS_CUSTOMER_CONFIRMED && order.ProductOrderStatus == utils.ORDERSTATUS_PRODUCT_CONFIRMED {
			order.OrderStatus = utils.ORDERSTATUS_CONFIRMED
			orderDetailsRepo.Store(od.Id, order)
			order.CallCustomer()
			order.CallProduct()
		}
	}

	c.Header("Content-Type", binding.MIMEPlain)
	c.String(http.StatusOK, "200 OK.")
}

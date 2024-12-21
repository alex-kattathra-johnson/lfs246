package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/alex-kattathra-johnson/lfs246/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var customerDetailsRepo sync.Map

func init() {
	data, _ := os.ReadFile("/data.json")
	var customers []utils.CustomerDetails
	json.Unmarshal(data, &customers)
	for _, c := range customers {
		customerDetailsRepo.Store(c.Id, c)
	}
}

func main() {
	r := gin.Default()
	r.POST("/customers/blockAmount", blockAmount)

	r.Run()
}

func blockAmount(c *gin.Context) {
	od := new(utils.OrderDetails)
	if err := c.Bind(od); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("customer-service :: Order Details :: %s", od)
	log.Infof("customer-service :: Order Status :: %s", od.OrderStatus)

	data, ok := customerDetailsRepo.Load(od.CustomerId)
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("customer not found"))
		return
	}

	customer := data.(utils.CustomerDetails)

	switch od.OrderStatus {
	case utils.ORDERSTATUS_NEW:
		customer.WalletAmountBlocked += int(od.Amount)
		customer.WalletAmount -= int(od.Amount)
		od.OrderStatus = utils.ORDERSTATUS_CUSTOMER_CONFIRMED
		customerDetailsRepo.Store(od.CustomerId, customer)
		od.ConfirmOrder()
	case utils.ORDERSTATUS_CONFIRMED:
		customer.WalletAmountBlocked -= int(od.Amount)
		customerDetailsRepo.Store(od.CustomerId, customer)
	}

	log.Infof("customer-service :: Order details :: %s", customer)
}

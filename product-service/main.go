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

var productDetailsRepo sync.Map

func init() {
	data, _ := os.ReadFile("/data.json")
	var products []utils.ProductDetails
	json.Unmarshal(data, &products)
	for _, p := range products {
		productDetailsRepo.Store(p.Id, p)
	}
}

func main() {
	r := gin.Default()
	r.POST("/products/block", block)

	r.Run()
}

func block(c *gin.Context) {
	od := new(utils.OrderDetails)
	if err := c.Bind(od); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Infof("product-service :: Order Details :: %s", od)
	log.Infof("product-service :: Order Status :: %s", od.OrderStatus)

	data, ok := productDetailsRepo.Load(od.ProductId)
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("product not found"))
		return
	}

	product := data.(utils.ProductDetails)

	switch od.OrderStatus {
	case utils.ORDERSTATUS_NEW:
		product.ProductBlocked += int(od.ProductCount)
		product.ProductAvailable -= int(od.ProductCount)
		od.OrderStatus = utils.ORDERSTATUS_PRODUCT_CONFIRMED
		productDetailsRepo.Store(od.ProductId, product)
		od.ConfirmOrder()
	case utils.ORDERSTATUS_CONFIRMED:
		product.ProductBlocked -= int(od.ProductCount)
		productDetailsRepo.Store(od.ProductId, product)
	}

	log.Infof("product-service :: Order details :: %s", product)
}

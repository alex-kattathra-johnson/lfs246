package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/alex-kattathra-johnson/lfs246/utils"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var productDetailsRepo *badger.DB

func init() {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatalf("could not create database: %s", err)
	}
	productDetailsRepo = db

	data, _ := os.ReadFile("/data.json")
	var products []utils.ProductDetails
	if err := json.Unmarshal(data, &products); err != nil {
		log.Fatalf("could not load products: %s", err)
	}

	if err := productDetailsRepo.Update(func(txn *badger.Txn) error {
		for _, p := range products {
			if _, err := txn.Get([]byte(p.Id)); err != nil {
				data, err := json.Marshal(p)
				if err != nil {
					return err
				}
				if err := txn.Set([]byte(p.Id), data); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("could not load products: %s", err)
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

	if err := productDetailsRepo.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(od.ProductId))
		if err != nil {
			c.AbortWithError(http.StatusNotFound, errors.New("product not found"))
			return nil
		}

		var data []byte
		if err := item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		}); err != nil {
			return err
		}

		var product utils.ProductDetails
		if err := json.Unmarshal(data, &product); err != nil {
			return err
		}

		switch od.OrderStatus {
		case utils.ORDERSTATUS_NEW:
			product.ProductBlocked += int(od.ProductCount)
			product.ProductAvailable -= int(od.ProductCount)
			od.OrderStatus = utils.ORDERSTATUS_PRODUCT_CONFIRMED
			od.SendTo("order")
		case utils.ORDERSTATUS_CONFIRMED:
			product.ProductBlocked -= int(od.ProductCount)
		}

		data, err = json.Marshal(product)
		if err != nil {
			return err
		}

		log.Infof("product-service :: Order details :: %s", product)
		return txn.Set([]byte(product.Id), data)
	}); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

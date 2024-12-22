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

var customerDetailsRepo *badger.DB

func init() {
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatalf("could not create database: %s", err)
	}
	customerDetailsRepo = db

	data, _ := os.ReadFile("/data.json")
	var customers []utils.CustomerDetails
	if err := json.Unmarshal(data, &customers); err != nil {
		log.Fatalf("could not load customers: %s", err)
	}

	if err := customerDetailsRepo.Update(func(txn *badger.Txn) error {
		for _, c := range customers {
			data, err := json.Marshal(c)
			if err != nil {
				return err
			}
			if err := txn.Set([]byte(c.Id), data); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("could not load customers: %s", err)
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

	if err := customerDetailsRepo.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(od.CustomerId))
		if err != nil {
			c.AbortWithError(http.StatusNotFound, errors.New("customer not found"))
			return nil
		}

		var data []byte
		if err := item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		}); err != nil {
			return err
		}

		var customer utils.CustomerDetails
		if err := json.Unmarshal(data, &customer); err != nil {
			return err
		}

		switch od.OrderStatus {
		case utils.ORDERSTATUS_NEW:
			customer.WalletAmountBlocked += int(od.Amount)
			customer.WalletAmount -= int(od.Amount)
			od.OrderStatus = utils.ORDERSTATUS_CUSTOMER_CONFIRMED
			od.ConfirmOrder()
		case utils.ORDERSTATUS_CONFIRMED:
			customer.WalletAmountBlocked -= int(od.Amount)
		}

		data, err = json.Marshal(customer)
		if err != nil {
			return err
		}

		log.Infof("customer-service :: Order details :: %s", customer)
		return txn.Set([]byte(customer.Id), data)
	}); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

package publish

import (
	"L0/order"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"log"
	"math/rand"
	"time"
)

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func randomDelivery() order.Delivery {
	return order.Delivery{
		Name:    randomString(10),
		Phone:   randomString(10),
		Zip:     randomString(6),
		City:    randomString(10),
		Address: randomString(20),
		Region:  randomString(10),
		Email:   randomString(15),
	}
}

func randomPayment() order.Payment {
	return order.Payment{
		Transaction:  randomString(12),
		RequestId:    randomString(8),
		Currency:     randomString(3),
		Provider:     randomString(10),
		Amount:       rand.Intn(1000),
		PaymentDt:    rand.Intn(1000000),
		Bank:         randomString(15),
		DeliveryCost: rand.Intn(100),
		GoodsTotal:   rand.Intn(500),
		CustomFee:    rand.Intn(50),
	}
}

func randomItems() []order.Item {
	numItems := rand.Intn(3) + 1
	items := make([]order.Item, numItems)
	for i := 0; i < numItems; i++ {
		item := order.Item{
			ChrtId:      rand.Intn(100),
			TrackNumber: randomString(8),
			Price:       rand.Intn(100),
			Rid:         randomString(10),
			Name:        randomString(12),
			Sale:        rand.Intn(50),
			Size:        randomString(5),
			TotalPrice:  rand.Intn(200),
			NmId:        rand.Intn(100),
			Brand:       randomString(10),
			Status:      rand.Intn(5),
		}
		items[i] = item
	}
	return items
}

func randomOrder() order.Order {
	return order.Order{
		OrderUID:          randomString(10),
		TrackNumber:       randomString(8),
		Entry:             randomString(5),
		Delivery:          randomDelivery(),
		Payment:           randomPayment(),
		Items:             randomItems(),
		Locale:            randomString(5),
		InternalSignature: randomString(16),
		CustomerID:        randomString(8),
		DeliveryService:   randomString(10),
		ShardKey:          randomString(5),
		SmID:              rand.Intn(1000),
		DateCreated:       time.Now(),
		OofShard:          randomString(5),
	}
}

func PublishGeneratedModels() error {
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://192.168.0.2:4222"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer sc.Close()

	rand.Seed(time.Now().UnixNano())
	numOrders := rand.Intn(10) + 1
	for i := 0; i < numOrders; i++ {
		order := randomOrder()

		file, err := json.Marshal(order)
		if err != nil {
			log.Println(err)
			return err
		}

		err = sc.Publish("my_channel", file)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return err
}

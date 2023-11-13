package main

import (
	"L0/order"
	"L0/publish"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

var dataCache = make(map[string]interface{})
var cacheMutex sync.Mutex

func addToCache(key string, value interface{}) {
	cacheMutex.Lock()
	dataCache[key] = value
	cacheMutex.Unlock()
}

func getFromCache(key string) (interface{}, bool) {
	cacheMutex.Lock()
	value, found := dataCache[key]
	cacheMutex.Unlock()
	return value, found
}

func main() {
	// Подключение к БД
	host := "192.168.0.3"
	port := 5432
	user := "valera"
	password := "123"
	dbname := "level0"
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer database.Close()

	// Извлечение данных из БД и загрузка в кэш
	rows, err := database.Query("SELECT id, data FROM orders")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	tmp1 := order.Order{}
	var id string
	var data string
	for rows.Next() {
		if err := rows.Scan(&id, &data); err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal([]byte(data), &tmp1)
		addToCache(fmt.Sprintf("%s", id), []byte(data))
	}

	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "subscriber", stan.NatsURL("nats://192.168.0.2:4222"))
	if err != nil {
		log.Fatalf("Ошибка при подключении к NATS Streaming: %v", err)
	}
	defer sc.Close()

	// Подписаться на сообщения в канале
	subscription, err := sc.Subscribe("my_channel", func(msg *stan.Msg) {
		var tmp order.Order

		err = json.Unmarshal(msg.Data, &tmp)
		if err != nil {
			log.Println(err)
			return
		}
		addToCache(tmp.OrderUID, msg.Data)

		_, err = database.Exec("INSERT INTO orders (id, data) VALUES ($1, $2)", tmp.OrderUID, msg.Data)
		if err != nil {
			log.Println(err)
			return
		}
	})
	if err != nil {
		log.Fatalf("Ошибка при подписке на канал: %v", err)
	}
	defer subscription.Unsubscribe()

	// Опубликовать сообщение в канал
	err = publish.PublishGeneratedModels()
	if err != nil {
		log.Fatalf("Ошибка при публикации сообщения: %v", err)
	}

	time.Sleep(4 * time.Second)

	// HTTP-сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Println(err)
		}
		err = html.Execute(w, nil)
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		requiredId := r.FormValue("find_id")
		jsonValue, found := getFromCache(requiredId)
		if found {
			var tmp order.Order
			byteSlice, _ := jsonValue.([]byte)
			err := json.Unmarshal(byteSlice, &tmp)
			if err != nil {
				log.Println("1")
				log.Fatal(err)
			}
			html, err := template.ParseFiles("templates/data.html")
			if err != nil {
				log.Println("2")
				log.Println(err)
			}
			err = html.Execute(w, tmp)
			if err != nil {
				log.Println("3")
				log.Println(err)
			}
		} else {
			html, err := template.ParseFiles("templates/404.html")
			if err != nil {
				log.Fatal(err)
			}
			err = html.Execute(w, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	// Запуск HTTP-сервера
	httpAddr := "192.168.0.4:8080"
	log.Printf("Starting HTTP server on %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatal(err)
	}
}

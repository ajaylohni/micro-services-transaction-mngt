package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/xid"
	"github.com/streadway/amqp"
)

//Message fields
type Message struct {
	Operation string `json:"operation,omitempty"`
	Table     string `json:"table,omitempty"`
	Database  string `json:"database,omitempty"`
	Schema    string `json:"schema,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Before    Before `json:"before,omitempty"`
	After     After  `json:"after,omitempty"`
}

//Before fields
type Before struct {
	ItemID   int    `json:"item_id,omitempty"`
	ItemName string `json:"item_name,omitempty"`
	ItemQty  int    `json:"item_quantity,omitempty"`
	Price    int    `json:"price,omitempty"`
	TID      string `json:"transaction_id,omitempty"`
}

//After fields
type After struct {
	ItemID   int    `json:"item_id,omitempty"`
	ItemName string `json:"item_name,omitempty"`
	ItemQty  int    `json:"item_quantity,omitempty"`
	Price    int    `json:"price,omitempty"`
	TID      string `json:"transaction_id,omitempty"`
}

var keyArray []string
var keyMap = make(map[string][]string)
var c = cache.New(500*time.Minute, 100*time.Minute)

func getTransactionID(w http.ResponseWriter, r *http.Request) {
	guid := xid.New()
	w.Write([]byte(guid.String()))
}

func finishTransaction(w http.ResponseWriter, r *http.Request) {
	tid := r.URL.Query()["tID"]
	tID := string(tid[0])
	c.Delete(tID)
	w.Write([]byte("Transaction has been successfully completed"))
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func getMessages() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	checkErr(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	checkErr(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	checkErr(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	checkErr(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for m := range msgs {
			getMsgValues(m.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func getMsgValues(msg []byte) {
	var m Message
	err := json.Unmarshal([]byte(msg), &m)
	ct, err := json.Marshal(&m)
	checkErr(err, "---json mapping failed")
	value := string(ct)
	key := m.After.TID
	flag := false
	for _, v := range keyArray {
		if v == key {
			flag = true
		}
	}
	if flag == false {
		keyArray = append(keyArray, key)
	}
	if _, ok := keyMap[key]; ok {
		keyMap[key] = append(keyMap[key], value)
	} else {
		keyMap[key] = append(keyMap[key], value)
	}
	for k, v := range keyMap {
		_, found := c.Get(k)
		if found {
			c.Delete(k)
			c.Set(k, v, cache.NoExpiration)
		} else {
			c.Set(k, v, cache.NoExpiration)
		}
	}

	foo, found := c.Get(key)
	if found {
		fmt.Println("\n--> this is the cache value for key ", key, "\n\n", foo)
	}
}

func main() {
	go getMessages()
	r := mux.NewRouter()
	r.HandleFunc("/getTransactionID", getTransactionID).Methods("GET")
	r.HandleFunc("/finishTransaction", finishTransaction).Methods("GET")
	fmt.Println(http.ListenAndServe(":9004", r))
}

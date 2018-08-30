package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var msg = []Message{}
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

func transactionFailed(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	m := Message{}
	defer func() { msg = []Message{} }()
	tid := r.URL.Query()["tID"]
	tID := string(tid[0])
	value, found := c.Get(tID)
	if found {
		c.Delete(tID)
		str := value.([]string)
		for _, v := range str {
			err := json.Unmarshal([]byte(v), &m)
			checkErr(err, "\njson unmarshal failed")
			msg = append(msg, m)
		}
		fmt.Printf("%v", msg)
		b, err := json.Marshal(msg)
		checkErr(err, "\njson unmarshal failed")
		url := "http://localhost:9002/rollBack"
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		status := resp.Status
		if status == "200 OK" {
			w.Write([]byte("Rollback successful"))
		} else {
			w.Write([]byte("Rollback failed!!!"))
		}
		checkErr(err, "rollback failed")
	} else {
		w.Write([]byte("transaction id not found"))
	}
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
}

func main() {
	go getMessages()
	r := mux.NewRouter()
	r.HandleFunc("/getTransactionID", getTransactionID).Methods("GET")
	r.HandleFunc("/finishTransaction", finishTransaction).Methods("GET")
	r.HandleFunc("/transactionFailed", transactionFailed).Methods("GET")
	fmt.Println(http.ListenAndServe(":9004", r))
}

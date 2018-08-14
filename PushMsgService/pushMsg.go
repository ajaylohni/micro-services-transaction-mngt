package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/streadway/amqp"
)

func createChannel() (*amqp.Channel, *amqp.Connection) {
	conn, err := amqp.Dial("amqp://guest:guest@RabbitMq:5672/")
	checkErr(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	checkErr(err, "Failed to open a channel")
	return ch, conn
}

func publishMsg(body []byte) {

	ch, conn := createChannel()

	q, err := ch.QueueDeclare(
		"FirstQueue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	checkErr(err, "Failed to declare a queue")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		})
	checkErr(err, "Failed to publish a message")
	defer conn.Close()
	defer ch.Close()
}

func pushMsg(w http.ResponseWriter, r *http.Request) {
	msg, _ := ioutil.ReadAll(r.Body)
	publishMsg(msg)
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/push", pushMsg).Methods("POST")
	http.ListenAndServe(":8006", r)
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	dbUser = "postgres"
	dbPass = "data"
	dbName = "test"
)

var db *sql.DB
var err error

func init() {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", dbinfo)
	checkErr(err, "Database connection failed")
}

func paymentStart(w http.ResponseWriter, r *http.Request) {
	tid := r.URL.Query()["tID"]
	cardno := r.URL.Query()["cardNo"]
	totalAmt := r.URL.Query()["totalAmount"]
	tID := string(tid[0])
	cardNo, err := strconv.Atoi(cardno[0])
	totalAmount, err := strconv.Atoi(totalAmt[0])
	checkErr(err, "reading params faield")
	var updatedAcc string
	balance, msg, err := getBalance(cardNo, totalAmount)
	if err == nil {
		sqlStatement := `update bank_table set balance=$1,transaction_id=$2 where card_no=$3 returning customer_name`
		err = db.QueryRow(sqlStatement, balance, tID, cardNo).Scan(&updatedAcc)
		if err == nil {
			w.WriteHeader(200)
			w.Write([]byte("Hi " + updatedAcc + " " + msg))
		}
	}
	checkErr(err, "Query err")
}

func getBalance(cardNo int, totalAmount int) (int, string, error) {
	var balance int
	var msg string
	sqlStatement := `select balance from bank_table where card_no=$1`
	err = db.QueryRow(sqlStatement, cardNo).Scan(&balance)
	if totalAmount > balance {
		msg = "tranasction failed, insufficient amount in your account"
	} else {
		balance = balance - totalAmount
		msg = "Payment successful"
	}
	return balance, msg, err
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/payment", paymentStart).Methods("GET")
	fmt.Println(http.ListenAndServe(":9003", r))
}

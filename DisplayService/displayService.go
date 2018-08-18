package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//conection parameters
const (
	host   = "Postgres"
	port   = 5432
	dbUser = "postgres"
	dbPass = "data"
	dbName = "postgres"
)

//Student Fields
type Student struct {
	USN         int    `json:"usn,omitempty"`
	Name        string `json:"name,omitempty"`
	Age         int    `json:"age,omitempty"`
	LastUpdated string `json:"lastUpdate,omitempty"`
}

func displayService(w http.ResponseWriter, r *http.Request) {

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err, "Database connection failed")
	defer db.Close()

	var usn, age int
	var name, lastUpdated string

	var jsonArray []string

	rows, err := db.Query("select * from students order by usn")
	checkErr(err, "Select statement failed")
	for rows.Next() {
		err := rows.Scan(&usn, &name, &age, &lastUpdated)
		row := jsonConvert(usn, name, age, lastUpdated)
		str := string(row)
		jsonArray = append(jsonArray, str)
		checkErr(err, "error")
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	temp := strings.Join(jsonArray, ",")
	temp = string("[") + temp + string("]")
	w.Write([]byte(temp))
}

func jsonConvert(usn int, name string, age int, lastUpdated string) []byte {
	m := Student{usn, name, age, lastUpdated}
	ct, err := json.Marshal(m)
	checkErr(err, "JSON conversion failed")
	return ct
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/display", displayService).Methods("GET")
	http.ListenAndServe(":8004", r)
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//conection parameters
const (
	host   = "Postgres"
	port   = 5432
	DbUser = "postgres"
	DbPass = "data"
	DbName = "postgres"
)

//Message Fields
type Message struct {
	USN      int      `json:"usn"`
	Before   *Before  `json:",omitempty"`
	After    After    `json:",omitempty"`
	Metadata Metadata `json:",omitempty"`
}

//Before Fields
type Before struct {
	USN         int    `json:"usn,omitempty"`
	Name        string `json:"name,omitempty"`
	Age         int    `json:"age,omitempty"`
	LastUpdated string `json:"lastUpdate,omitempty"`
}

//After Fields
type After struct {
	USN         int    `json:"usn,omitempty"`
	Name        string `json:"name,omitempty"`
	Age         int    `json:"age,omitempty"`
	LastUpdated string `json:"lastUpdate,omitempty"`
}

//Metadata Fields
type Metadata struct {
	DatabaseName string `json:"database,omitempty"`
	TableName    string `json:"table,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

func deleteService(w http.ResponseWriter, r *http.Request) {

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, DbUser, DbPass, DbName)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err, "Database Connection failed")
	defer db.Close()

	id := mux.Vars(r)

	temp := id["id"]
	usn, err := strconv.Atoi(temp)
	checkErr(err, "string to integer convesion failed")
	msg := scanRow(db, usn)

	res, err := db.Exec("delete from students where usn=$1", usn)
	checkErr(err, "Delete statement failed")
	affect, err := res.RowsAffected()
	checkErr(err, "Row affect failed")
	if affect == 1 {
		req, err := http.NewRequest("POST", "http://PushMsg:8006/push", bytes.NewBuffer(msg))
		req.Close = true
		req.Header.Set("Content-Type", "text/plain")
		client := &http.Client{}
		resp, err := client.Do(req)
		checkErr(err, "Request to push message service failed")
		defer resp.Body.Close()
		w.Write([]byte("Record deleted successfully"))
	} else {
		w.Write([]byte("Delete Operation failed"))
	}

}

func scanRow(db *sql.DB, tag int) []byte {
	var usn, age int
	var name, lastUpdated string
	var msg []byte

	rows, err := db.Query("select * from students where usn=$1", tag)
	checkErr(err, "select statement failed")
	for rows.Next() {
		err := rows.Scan(&usn, &name, &age, &lastUpdated)
		msg = jsonConvert(usn, name, age, lastUpdated)
		checkErr(err, "scan row failed")
	}
	return msg
}

func jsonConvert(usn int, name string, age int, lastUpdated string) []byte {
	var m Message
	t := time.Now()
	m = Message{usn, &Before{usn, name, age, t.Format("2006-01-02 15:04:05")}, After{}, Metadata{"postgres", "students", "Delete"}}
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
	corsObj := handlers.AllowedOrigins([]string{"*"})
	corsObj1 := handlers.AllowedMethods([]string{"DELETE"})
	r.HandleFunc("/delete/{id}", deleteService).Methods("DELETE")
	http.ListenAndServe(":8002", handlers.CORS(corsObj, corsObj1)(r))
}

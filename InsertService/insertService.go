package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

//Student fields
type Student struct {
	USN  int    `json:"usn"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

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

func insertService(w http.ResponseWriter, r *http.Request) {
	s := Student{}
	json.NewDecoder(r.Body).Decode(&s)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, DbUser, DbPass, DbName)
	db, _ := sql.Open("postgres", dbinfo)
	defer db.Close()

	t := time.Now()
	var lastInsertID int
	sqlStatement := `INSERT INTO students(usn,name,age,time) VALUES($1,$2,$3,$4) returning usn`
	err := db.QueryRow(sqlStatement, s.USN, s.Name, s.Age, t.Format("2006-01-02 15:04:05")).Scan(&lastInsertID)

	if lastInsertID == s.USN {
		scanRow(db, s.USN)
		w.Write([]byte("Record inserted successfully"))
	} else {
		w.Write([]byte("Failed"))
	}
	checkErr(err, "Insert Service Failed")
}

func scanRow(db *sql.DB, tag int) {
	var usn, age int
	var name, lastUpdated string

	rows, _ := db.Query("select * from students where usn=$1", tag)
	for rows.Next() {
		err := rows.Scan(&usn, &name, &age, &lastUpdated)
		msg := jsonConvert(usn, name, age, lastUpdated)
		req, err := http.NewRequest("POST", "http://PushMsg:8006/push", bytes.NewBuffer(msg))
		req.Close = true
		req.Header.Set("Content-Type", "text/plain")
		client := &http.Client{}
		resp, err := client.Do(req)
		checkErr(err, "scan row block failed")
		defer resp.Body.Close()
	}
}

func jsonConvert(usn int, name string, age int, lastUpdated string) []byte {
	var m Message
	m = Message{usn, &Before{}, After{usn, name, age, lastUpdated}, Metadata{"postgres", "students", "Insert"}}
	ct, err := json.Marshal(m)
	checkErr(err, "JSON convert block")
	return ct
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/insert", insertService).Methods("POST")
	http.ListenAndServe(":8001", r)
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/gorilla/mux"
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

var tableExists bool

func init() {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err, "Database connection failed")

	err = db.QueryRow("SELECT EXISTS (SELECT * FROM information_schema.tables WHERE table_name = 'students')").Scan(&tableExists)
	checkErr(err, "Checking table existing failed")
	if tableExists != true {
		stmt, err := db.Prepare("create table students(usn integer primary key,name text,age integer,time timestamp without time zone)")
		checkErr(err, "Create table failed")
		_, err = stmt.Exec()
	} else {
		pushAllmsg(db)
	}
	checkErr(err, "init block failed")
}

func pushAllmsg(db *sql.DB) {
	var usn, age int
	var name, lastUpdated string

	rows, err := db.Query("select * from students order by usn")
	checkErr(err, "Select statement failed")
	for rows.Next() {
		err := rows.Scan(&usn, &name, &age, &lastUpdated)
		msg := jsonConvert(usn, name, age, lastUpdated)
		req, err := http.NewRequest("POST", "http://PushMsg:8006/push", bytes.NewBuffer(msg))
		req.Close = true
		req.Header.Set("Content-Type", "text/plain")
		client := &http.Client{}
		resp, err := client.Do(req)
		checkErr(err, "Pushing message failed")
		defer resp.Body.Close()
	}
	defer db.Close()
}

func jsonConvert(usn int, name string, age int, lastUpdated string) []byte {
	var m Message
	m = Message{usn, &Before{}, After{usn, name, age, lastUpdated}, Metadata{"postgres", "students", "initial Snapshot"}}
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
	fmt.Println("initial snapshot")
}

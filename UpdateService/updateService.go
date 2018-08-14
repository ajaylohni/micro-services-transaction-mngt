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

//Connection parameters
const (
	host   = "Postgres"
	port   = 5432
	dbUser = "postgres"
	dbPass = "data"
	dbName = "postgres"
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

func updateService(w http.ResponseWriter, r *http.Request) {
	s := Student{}
	json.NewDecoder(r.Body).Decode(&s)

	id := mux.Vars(r)
	temp := id["id"]
	usn, _ := strconv.Atoi(temp)

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, _ := sql.Open("postgres", dbinfo)
	defer db.Close()

	t := time.Now()

	usn1, name1, age1, lastUpdated1 := rowScan(db, s.USN)

	var lastUpdateID int
	sqlStatement := `update students set usn=$1,name=$2,age=$3,time=$4 where usn=$5 returning usn`
	err := db.QueryRow(sqlStatement, s.USN, s.Name, s.Age, t.Format("2006-01-02 15:04:05"), usn).Scan(&lastUpdateID)
	checkErr(err, "Query err")

	usn2, name2, age2, lastUpdated2 := rowScan(db, s.USN)

	if lastUpdateID == usn {
		msg := jsonConvert(usn1, name1, age1, lastUpdated1, usn2, name2, age2, lastUpdated2)
		req, err := http.NewRequest("POST", "http://PushMsg:8006/push", bytes.NewBuffer(msg))
		req.Close = true
		req.Header.Set("Content-Type", "text/plain")
		client := &http.Client{}
		resp, err := client.Do(req)
		checkErr(err, "scan row failed")
		w.Write([]byte("Record updated successfully"))
		defer resp.Body.Close()
	} else {
		w.Write([]byte("Update Operation failed"))
	}
}

func rowScan(db *sql.DB, tag int) (int, string, int, string) {
	var usn, age int
	var name, lastUpdated string

	rows, _ := db.Query("select * from students where usn=$1", tag)

	for rows.Next() {
		err := rows.Scan(&usn, &name, &age, &lastUpdated)
		checkErr(err, "Row scan failed")
	}
	return usn, name, age, lastUpdated
}

func jsonConvert(usn int, name string, age int, lastUpdated string, usn1 int, name1 string, age1 int, lastUpdated1 string) []byte {
	m := &Message{usn, &Before{usn, name, age, lastUpdated}, After{usn1, name1, age1, lastUpdated1}, Metadata{"postgres", "students", "Update"}}
	ct, err := json.Marshal(m)
	checkErr(err, "Failed at conveting into JSON")
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
	corsObj1 := handlers.AllowedMethods([]string{"PUT"})
	r.HandleFunc("/update/{id}", updateService).Methods("PUT")
	http.ListenAndServe(":8003", handlers.CORS(corsObj, corsObj1)(r))
}

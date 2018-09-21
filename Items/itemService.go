package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host   = "Postgres"
	port   = 5432
	dbUser = "postgres"
	dbPass = "data"
	dbName = "test"
)

//Items struct fileds
type Items struct {
	ItemID       int    `json:"item_id,omitempty"`
	ItemName     string `json:"item_name,omitempty"`
	ItemQuantity int    `json:"item_qty,omitempty"`
	ItemPrice    int    `json:"item_price,omitempty"`
}

//OrderList fields
type OrderList struct {
	TID      string  `json:"transaction_id,omitempty"`
	CartItem []Items `json:"cart_items,omitempty"`
}

//UpdatedValues fields
type UpdatedValues struct {
	ItemID       int `json:"item_id,omitempty"`
	ItemQuantity int `json:"item_qty,omitempty"`
}

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

var db *sql.DB
var err error

var items []byte
var allItems []Items

func init() {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", dbinfo)
	checkErr(err, "\n-> Database connection failed")
}

func getItems(w http.ResponseWriter, r *http.Request) {
	var item Items
	var discarted interface{}
	var jsonArray []string

	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Getting items from database failed", r)
		}
	}()

	rows, err := db.Query("select * from item_table order by item_id")
	checkErr(err, "\n-> Getting data from database failed")
	for rows.Next() {
		err := rows.Scan(&item.ItemID, &item.ItemName, &item.ItemQuantity, &item.ItemPrice, &discarted)
		row := jsonConvert(item.ItemID, item.ItemName, item.ItemQuantity, item.ItemPrice)
		str := string(row)
		jsonArray = append(jsonArray, str)
		checkErr(err, "\n-> Mapping row to struct failed")
	}
	temp := strings.Join(jsonArray, ",")
	temp = string("[") + temp + string("]")
	items = []byte(temp)
	w.Write([]byte(temp))
}

func jsonConvert(itemID int, itemName string, itemQty int, itemPrc int) []byte {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> JSON marshal failed", r)
		}
	}()
	i := Items{itemID, itemName, itemQty, itemPrc}
	ct, err := json.Marshal(i)
	checkErr(err, "\n-> JSON marshal failed in jsonConvert")
	return ct
}

func updateItems(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> updating items in database failed", r)
		}
	}()
	var updatedItemsList []int
	var cartJSON OrderList
	err := json.NewDecoder(r.Body).Decode(&cartJSON)
	checkErr(err, "\n-> failed to decode request json")
	last := getUpdatedValues(cartJSON)
	fmt.Println(last)
	for _, v := range last {
		var updatedItem int
		sqlStatement := `update item_table set item_quantity=$1,transaction_id=$2 where item_id=$3 returning item_id`
		err = db.QueryRow(sqlStatement, v.ItemQuantity, cartJSON.TID, v.ItemID).Scan(&updatedItem)
		if err == nil {
			updatedItemsList = append(updatedItemsList, updatedItem)
		}
		checkErr(err, "\n-> Update items to database failed")
	}
	s, err := json.Marshal(updatedItemsList)
	checkErr(err, "\n-> JSON Marshal failed")
	w.WriteHeader(200)
	w.Write([]byte("Updated Items : " + string(s)))
}

func getUpdatedValues(cartJSON OrderList) []UpdatedValues {
	var lastUpdate []UpdatedValues
	finalValues := UpdatedValues{}
	err = json.Unmarshal(items, &allItems)
	checkErr(err, "\n-> JSON unmarshal failed at getting updated values")
	for _, vo := range allItems {
		for _, vi := range cartJSON.CartItem {
			if vo.ItemID == vi.ItemID {
				finalValues.ItemID = vi.ItemID
				calc := vo.ItemQuantity - vi.ItemQuantity
				finalValues.ItemQuantity = calc
				lastUpdate = append(lastUpdate, finalValues)
			}
		}
	}
	return lastUpdate
}

func rollBack(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Rollback failed", r)
		}
	}()
	rollbackJSON := []Message{}
	contents, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(contents, &rollbackJSON)
	checkErr(err, "\n-> JSON unmarshal failed at rollback")
	var itemID interface{}
	for _, v := range rollbackJSON {
		switch v.Operation {
		case "INSERT":
			stmt := `delete from item_table where item_id=$1 returning item_id`
			err := db.QueryRow(stmt, v.After.ItemID).Scan(&itemID)
			fmt.Println("Rollback on Item id : ", itemID)
			/*this is for test */
			fmt.Printf("\nThe Item values are : %+v", v)
			checkErr(err, "Query err")
		case "DELETE":
			stmt := `insert into item_table values($1,$2,$3,$4,null) returning item_id`
			err := db.QueryRow(stmt, v.Before.ItemID, v.Before.ItemName, v.Before.ItemQty, v.Before.Price).Scan(&itemID)
			fmt.Println("Rollback on Item id : ", itemID)
			/*this is for test */
			fmt.Printf("\nThe Item values are : %+v", v)
			checkErr(err, "Query err")
		case "UPDATE":
			stmt := `update item_table set item_quantity=$1 where item_id=$2 returning item_id`
			err := db.QueryRow(stmt, v.Before.ItemQty, v.Before.ItemID).Scan(&itemID)
			fmt.Println("Rollback on Item id : ", itemID)
			/*this is for test */
			fmt.Printf("\nThe Item values are : %+v", v)
			checkErr(err, "Query err")
		default:
			fmt.Println("Default executed")
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nRollback successful"))
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/getItems", getItems).Methods("GET")
	r.HandleFunc("/updateItems", updateItems).Methods("PUT")
	r.HandleFunc("/rollBack", rollBack).Methods("GET")
	fmt.Println(http.ListenAndServe(":9002", r))
}

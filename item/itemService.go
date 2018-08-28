package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

var db *sql.DB
var err error

var items []byte
var allItems []Items

func init() {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", dbinfo)
	checkErr(err, "Database connection failed")
}

func getItems(w http.ResponseWriter, r *http.Request) {
	var item Items
	var discarted interface{}
	var jsonArray []string
	rows, err := db.Query("select * from item_table order by item_id")
	checkErr(err, "Select statement failed")
	for rows.Next() {
		err := rows.Scan(&item.ItemID, &item.ItemName, &item.ItemQuantity, &item.ItemPrice, &discarted)
		row := jsonConvert(item.ItemID, item.ItemName, item.ItemQuantity, item.ItemPrice)
		str := string(row)
		jsonArray = append(jsonArray, str)
		checkErr(err, "Mapping to struct failed")
	}
	temp := strings.Join(jsonArray, ",")
	temp = string("[") + temp + string("]")
	items = []byte(temp)
	w.Write([]byte(temp))
}

func jsonConvert(itemID int, itemName string, itemQty int, itemPrc int) []byte {
	i := Items{itemID, itemName, itemQty, itemPrc}
	ct, err := json.Marshal(i)
	checkErr(err, "JSON convert block")
	return ct
}

func updateItems(w http.ResponseWriter, r *http.Request) {
	var updatedItemsList []int
	var cartJSON OrderList
	err := json.NewDecoder(r.Body).Decode(&cartJSON)
	checkErr(err, "failed to decode request json")
	last := getUpdatedValues(cartJSON)
	fmt.Println(last)
	for _, v := range last {
		var updatedItem int
		sqlStatement := `update item_table set item_quantity=$1,transaction_id=$2 where item_id=$3 returning item_id`
		err = db.QueryRow(sqlStatement, v.ItemQuantity, cartJSON.TID, v.ItemID).Scan(&updatedItem)
		if err == nil {
			updatedItemsList = append(updatedItemsList, updatedItem)
		}
		checkErr(err, "Query err")
	}
	s, err := json.Marshal(updatedItemsList)
	checkErr(err, "JSON Marshal failed")
	w.WriteHeader(200)
	w.Write([]byte("Updated Items : " + string(s)))
}

func getUpdatedValues(cartJSON OrderList) []UpdatedValues {
	var lastUpdate []UpdatedValues
	finalValues := UpdatedValues{}
	err = json.Unmarshal(items, &allItems)
	checkErr(err, "json problem")
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

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/getItems", getItems).Methods("GET")
	r.HandleFunc("/updateItems", updateItems).Methods("PUT")
	fmt.Println(http.ListenAndServe(":9002", r))
}

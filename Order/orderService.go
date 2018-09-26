package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"text/tabwriter"
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

var itemArray, qtyArray, priceArray []int
var itemNames []string
var key, maxQty int
var flag = false

var totalItems = []Items{}
var cartItems = []Items{}
var items = []Items{}
var totalPrice = 0
var ch = "y"
var client = &http.Client{}

func init() {
	clearScreen()
}

func getItems() {
	var item []Items
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> getItems failed", r)
		}
	}()
	response, err := http.Get("http://Items:9002/getItems")
	checkErr(err, "\n-> Getting items from Item service Failed")
	contents, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(contents, &item)
	checkErr(err, "\n-> Json unmarshal of items failed")
	defer response.Body.Close()
	totalItems = item
	displayItems(item, "Item")
}

func displayItems(items []Items, title string) {
	tw := tabwriter.NewWriter(os.Stdout, 10, 8, 0, '\t', 0)
	fmt.Println("\n\t\t\t", title, " Table")
	printLine()
	fmt.Fprintf(tw, "| Item ID\tItem Name\tQuantity\tPrice\t|\n")
	tw.Flush()
	printLine()
	for _, v := range items {
		if title == "Item" {
			itemArray = append(itemArray, v.ItemID)
			itemNames = append(itemNames, v.ItemName)
			qtyArray = append(qtyArray, v.ItemQuantity)
			priceArray = append(priceArray, v.ItemPrice)
		}
		fmt.Fprintf(tw, "| %v\t%v\t%v\t%v\t|\n", v.ItemID, v.ItemName, v.ItemQuantity, v.ItemPrice)
		tw.Flush()
	}
	printLine()
	if title == "Cart" {
		fmt.Fprintf(tw, "| Total Amount\t \t \t%v\t|\n", totalPrice)
		tw.Flush()
		printLine()
	}
}

func getOrder() {
	for ch == "y" {
		var itemID, itemQuantity int
		fmt.Print("\nEnter Item ID : ")
		fmt.Scanln(&itemID)
		fmt.Print("Enter Item Quantity : ")
		fmt.Scanln(&itemQuantity)
		flag = checkItem(itemID)
		flag = checkQuantity(itemQuantity)
		if flag == true {
			fmt.Println("Your order is...")
			itemPrice := priceArray[key] * itemQuantity
			totalPrice += itemPrice
			cart(itemID, itemNames[key], itemQuantity, itemPrice)
		} else {
			fmt.Println("Choose currect item or Quantity is more than available")
			break
		}
		fmt.Print("Do you want to order more items(y/n) : ")
		fmt.Scanln(&ch)
		if ch == "y" {
			clearScreen()
			displayItems(totalItems, "Item")
			displayItems(cartItems, "Cart")
		} else {
			var c string
			fmt.Print("Total amount is : ", totalPrice, "/- Press y to proceed for payment or n for exit...  ")
			fmt.Scanln(&c)
			if c == "y" {
				tID := getUUID()
				if tID != "" {
					var cardNo int
					var status string
					status = updateItems(tID, cartItems)
					fmt.Print("Enter your card No : ")
					fmt.Scanln(&cardNo)
					status = payment(tID, cardNo, totalPrice)
					if status == "200 OK" {
						finishTransaction(tID)
					} else {
						fmt.Println("calling transaction rollback")
						transactionFailed(tID)
					}
				} else {
					fmt.Println("\n-> Failed to get transaction ID")
				}
			} else {
				os.Exit(1)
			}
		}
	}
}

func cart(itemID int, itemName string, itemQty int, itemPrice int) {
	item := Items{itemID, itemName, itemQty, itemPrice}
	var temp = false
	for k, v := range items {
		if v.ItemID == itemID {
			value := &items[k]
			value.ItemQuantity += itemQty
			value.ItemPrice += itemPrice
			temp = true
			break
		}
	}
	if temp == false {
		items = append(items, item)
	}
	cartItems = items
	displayItems(items, "Cart")
}

func updateItems(tID string, cartItem []Items) string {
	cartJSON := OrderList{tID, cartItem}
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> updateItems failed", r)
		}
	}()
	url := "http://Items:9002/updateItems"
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&cartJSON)
	req, err := http.NewRequest("PUT", url, b)
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	status := resp.Status
	checkErr(err, "\n-> Updating items failed")
	return status
}

func payment(tID string, cardNo int, totalAmount int) string {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Payment failed", r)
		}
	}()
	url := "http://Payment:9003/payment"
	req, err := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("tID", tID)
	q.Add("cardNo", strconv.Itoa(cardNo))
	q.Add("totalAmount", strconv.Itoa(totalAmount))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	status := resp.Status
	fmt.Println(string(body))
	checkErr(err, "\n-> Payment service call failed")
	return status
}

func getUUID() string {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Getting UID failed", r)
		}
	}()
	url := "http://Transaction:9004/getTransactionID"
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	checkErr(err, "\n-> Failed at getting transaction ID")
	return string(body)
}

func finishTransaction(tID string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Finishing Transaction failed:", r)
		}
	}()
	url := "http://Transaction:9004/finishTransaction"
	req, err := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("tID", tID)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	checkErr(err, "Fail to finish")
}

func transactionFailed(tID string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("\n-> Rollback call failed", r)
		}
	}()
	url := "http://Transaction:9004/transactionFailed"
	req, err := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("tID", tID)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	checkErr(err, "\n-> Fail to rollback")
}

func checkItem(itemID int) bool {
	for k, v := range itemArray {
		if v == itemID {
			flag = true
			key = k
		}
	}
	return flag
}

func checkQuantity(iQ int) bool {
	if iQ > qtyArray[key] {
		flag = false
	}
	qtyArray[key] -= iQ
	return flag
}

func printLine() {
	for i := 0; i < 65; i++ {
		fmt.Print("-")
	}
	fmt.Println()
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func checkErr(err error, text string) {
	if err != nil {
		log.Fatal(err, text)
	}
}

func main() {
	fmt.Println("\nThis is the order service...")
	getItems()
	getOrder()
	/* r := mux.NewRouter()
	r.HandleFunc("/getOrder", getOrder).Methods("GET")
	fmt.Println(http.ListenAndServe(":9001", r)) */
}

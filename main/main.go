package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/senpainikolay/FoodOrderingService/structs"
)

var res structs.Restaurants
var conf = structs.GetConf()

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/register", RegisterRestaurant).Methods("POST")
	r.HandleFunc("/order", ClientOrderPost).Methods("POST")
	r.HandleFunc("/menu", GetMenu).Methods("GET")

	http.ListenAndServe(":"+conf.Port, r)
}

func RegisterRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var resReg structs.RegisterPayload
	err := json.NewDecoder(r.Body).Decode(&resReg)
	if err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}
	res.Mutex.Lock()
	res.Info = append(res.Info, resReg)
	res.Mutex.Unlock()
	fmt.Fprintf(w, "Restaurant id %v have been succesfully registered at Orders Manger", resReg.RestaurnatId)

}

func ClientOrderPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var ords structs.Orders
	err := json.NewDecoder(r.Body).Decode(&ords)
	if err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}

	var wg sync.WaitGroup
	wg.Add(len(ords.Orders))
	var clientResponse structs.ClientResponse
	clientResponse.OrderId = ords.ClientId
	// Send order to Dining-Hall and wait for response
	for i := range ords.Orders {
		idx := i
		go func() {
			clientResponse.Orders = append(clientResponse.Orders,
				SendOrderToDH(
					&structs.OrderToDiningHall{
						Items:       ords.Orders[idx].Items,
						Priority:    ords.Orders[idx].Priority,
						MaxWait:     ords.Orders[idx].MaxWait,
						CreatedTime: ords.Orders[idx].CreatedTime,
					}, res.Info[GetIndexForResId(ords.Orders[idx].RestaurantId)].Address))
			wg.Done()

		}()

	}
	wg.Wait()
	resp, _ := json.Marshal(&clientResponse)
	fmt.Fprint(w, string(resp))

}

func GetIndexForResId(id int) int {
	res.Mutex.Lock()

	for i, restaurant := range res.Info {
		if restaurant.RestaurnatId == id {
			res.Mutex.Unlock()
			return i
		}
	}
	log.Panicln("Couldnt find Restaurant Id in the storage at Food Ordering!!!")
	res.Mutex.Unlock()
	return -1
}

func SendOrderToDH(ord *structs.OrderToDiningHall, address string) structs.OMResponse {

	postBody, _ := json.Marshal(ord)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://"+address+"/v2/order", "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	var clientResponse structs.ClientOrderResponse
	if err := json.Unmarshal([]byte(body), &clientResponse); err != nil {
		panic(err)
	}

	resp.Body.Close()
	return structs.OMResponse{
		OrderId:              clientResponse.OrderId,
		RestaurantId:         clientResponse.RestaurantId,
		EstimatedWaitingTime: clientResponse.EstimatedWaitingTime,
		CreatedTime:          clientResponse.CreatedTime,
		RegisteredTime:       clientResponse.RegisteredTime,
		RestaurantAddress:    address,
	}

}

func GetMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var resData []structs.RestaurantData
	res.Mutex.Lock()
	for i := range res.Info {
		resData = append(resData, structs.RestaurantData{
			RestaurnatId: res.Info[i].RestaurnatId,
			Name:         res.Info[i].Name,
			MenuItems:    res.Info[i].MenuItems,
			Menu:         res.Info[i].Menu,
			Rating:       res.Info[i].Rating,
		})
	}
	sendMenu := structs.MenuGet{
		Restaurants:     len(res.Info),
		RestaurantsData: resData,
	}
	res.Mutex.Unlock()
	resp, _ := json.Marshal(&sendMenu)
	fmt.Fprint(w, string(resp))

}

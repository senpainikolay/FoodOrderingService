package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/senpainikolay/FoodOrderingService/structs"
)

var res structs.Restaurants

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/register", RegisterRestaurant).Methods("POST")
	r.HandleFunc("/order", ClientOrderPost).Methods("POST")
	r.HandleFunc("/menu", GetMenu).Methods("GET")

	http.ListenAndServe(":5000", r)

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
	log.Println(res.Info[len(res.Info)-1])
	fmt.Fprintf(w, "Your restaurant have been succesfully registered at Orders Manger")

}

func ClientOrderPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "KEK")

}

func GetMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var resData []structs.RestaurantData
	res.Mutex.Lock()
	for i := range res.Info {
		resData = append(resData, res.Info[i].RestaurantData)
	}
	sendMenu := structs.MenuGet{
		Restaurants:     len(res.Info),
		RestaurantsData: resData,
	}
	res.Mutex.Unlock()
	resp, _ := json.Marshal(&sendMenu)
	fmt.Fprint(w, string(resp))

}

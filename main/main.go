package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/senpainikolay/FoodOrderingService/structs"
)

var res structs.Restaurants
var conf = structs.GetConf()
var rating = GetRatingStruct()

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/register", RegisterRestaurant).Methods("POST")
	r.HandleFunc("/order", ClientOrderPost).Methods("POST")
	r.HandleFunc("/menu", GetMenu).Methods("GET")
	r.HandleFunc("/rating", ClientRatingPost).Methods("POST")

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
func ClientRatingPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// 204 status code
	w.WriteHeader(http.StatusNoContent)
	var clientRatingPost structs.ClientPostRating
	err := json.NewDecoder(r.Body).Decode(&clientRatingPost)
	if err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}

	var avg_rating float64
	avg_rating = 0.0
	var wg sync.WaitGroup
	wg.Add(len(clientRatingPost.Orders))
	for _, or := range clientRatingPost.Orders {
		ord := or
		go func() {
			order := structs.RestaurantRatingPayload{
				OrderId:              ord.OrderId,
				Rating:               ord.Rating,
				EstimatedWaitingTime: ord.EstimatedWaitingTime,
				WaitingTime:          ord.WaitingTime,
			}
			add := res.Info[GetIndexForResId(ord.RestaurantId)].Address

			avg_rating += SendRatingPaylodToRes(&order, add)
			wg.Done()
		}()
	}
	wg.Wait()

	avg_fn := avg_rating / float64(len(clientRatingPost.Orders))
	rating.Add(avg_fn)

}

func SendRatingPaylodToRes(py *structs.RestaurantRatingPayload, address string) float64 {

	postBody, _ := json.Marshal(py)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://"+address+"/v2/rating", "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}
	var resResp structs.RestaurantRatingResponse
	if err := json.Unmarshal([]byte(body), &resResp); err != nil {
		panic(err)
	}

	resp.Body.Close()
	return resResp.RestaurantAvgRating

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
	deadResturantsId := make([]int, 0)

	// Send order to Dining-Hall and wait for response
	for i := range ords.Orders {
		idx := i
		go func() {
			// in case Res have been removed
			index := GetIndexForResId(ords.Orders[idx].RestaurantId)
			if index == -1 {
				wg.Done()
				return
			}
			addr := res.Info[index].Address
			// GET req to dining hall -> kitchen and returns the current number of orders being preparing
			BusyIndex, err := GetBusyIndex(addr)
			// Handle dead request to Restaurants
			if err != nil {
				deadResturantsId = append(deadResturantsId, ords.Orders[idx].RestaurantId)
				wg.Done()
				return
			}
			// If everythin is fine and restaurant is not busy
			if BusyIndex == 0 {

				clientResponse.Orders = append(clientResponse.Orders,
					SendOrderToDH(
						&structs.OrderToDiningHall{
							Items:       ords.Orders[idx].Items,
							Priority:    ords.Orders[idx].Priority,
							MaxWait:     ords.Orders[idx].MaxWait,
							CreatedTime: ords.Orders[idx].CreatedTime,
						}, addr))

			}
			wg.Done()

		}()

	}
	wg.Wait()
	// Remove Res
	for len(deadResturantsId) != 0 {
		idx := GetIndexForResId(deadResturantsId[0])
		res.Mutex.Lock()
		res.Info = remove(res.Info, idx)
		deadResturantsId = pop(deadResturantsId)
		res.Mutex.Unlock()
	}
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

func GetBusyIndex(address string) (int, error) {

	resp, err := http.Get("http://" + address + "/getOrderStatus")
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	ok, _ := strconv.Atoi(string(body))

	return ok, nil
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

func remove(slice []structs.RegisterPayload, s int) []structs.RegisterPayload {
	return append(slice[:s], slice[s+1:]...)
}

func pop(slice []int) []int {
	if len(slice) == 1 {
		return []int{}
	}
	return slice[1:]
}

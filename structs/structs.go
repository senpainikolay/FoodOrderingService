package structs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Food struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	PreparationTime  int    `json:"preparation_time"`
	Complexity       int    `json:"complexity"`
	CookingApparatus string `json:"cooking_apparatus"`
}
type RestaurantData struct {
	RestaurnatId int     `json:"restaurant_id"`
	Name         string  `json:"name"`
	MenuItems    int     `json:"menu_items"`
	Menu         []Food  `json:"menu"`
	Rating       float64 `json:"rating"`
}

type Restaurants struct {
	Info  []RegisterPayload
	Mutex sync.Mutex
}

type RegisterPayload struct {
	RestaurnatId int     `json:"restaurant_id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	MenuItems    int     `json:"menu_items"`
	Menu         []Food  `json:"menu"`
	Rating       float64 `json:"rating"`
}

type MenuGet struct {
	Restaurants     int              `json:"restaurants"`
	RestaurantsData []RestaurantData `json:"restaurants_data"`
}

type Order struct {
	RestaurantId int `json:"restaurant_id"`
	OrderToDiningHall
}

type OrderToDiningHall struct {
	Items       []int   `json:"items"`
	Priority    int     `json:"priority"`
	MaxWait     float64 `json:"max_wait"`
	CreatedTime int64   `json:"created_time"`
}

type Orders struct {
	ClientId int     `json:"client_id"`
	Orders   []Order `json:"orders"`
}

type ClientOrderResponse struct {
	RestaurantId         int     `json:"restaurant_id"`
	OrderId              int     `json:"order_id"`
	EstimatedWaitingTime float64 `json:"estimated_waiting_time"`
	CreatedTime          int64   `json:"created_time"`
	RegisteredTime       int64   `json:"registered_time"`
}

type OMResponse struct {
	RestaurantId         int     `json:"restaurant_id"`
	RestaurantAddress    string  `json:"restaurant_address"`
	OrderId              int     `json:"order_id"`
	EstimatedWaitingTime float64 `json:"estimated_waiting_time"`
	CreatedTime          int64   `json:"created_time"`
	RegisteredTime       int64   `json:"registered_time"`
}

type ClientResponse struct {
	OrderId int          `json:"order_id"`
	Orders  []OMResponse `json:"orders"`
}

type ClientPostRating struct {
	ClientId int           `json:"client_id"`
	OrderId  int           `json:"order_id"`
	Orders   []RatingOrder `json:"orders"`
}

type RatingOrder struct {
	RestaurantId         int     `json:"restaurant_id"`
	OrderId              int     `json:"order_id"`
	Rating               int     `json:"rating"`
	EstimatedWaitingTime float64 `json:"estimated_waiting_time"`
	WaitingTime          int     `json:"waiting_time"`
}

type Conf struct {
	Port string `json:"port"`
}

type RestaurantRatingPayload struct {
	OrderId              int     `json:"order_id"`
	Rating               int     `json:"rating"`
	EstimatedWaitingTime float64 `json:"estimated_waiting_time"`
	WaitingTime          int     `json:"waiting_time"`
}

type RestaurantRatingResponse struct {
	RestaurantId        int     `json:"restaurant_id"`
	RestaurantAvgRating float64 `json:"restaurant_avg_rating"`
	PreparedOrders      int     `json:"prepared_orders"`
}

func GetConf() *Conf {
	jsonFile, err := os.Open("configurations/Conf.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var conf Conf
	json.Unmarshal(byteValue, &conf)
	return &conf

}

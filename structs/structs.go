package structs

import "sync"

type Food struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	PreparationTime  int    `json:"preparation_time"`
	Complexity       int    `json:"complexity"`
	CookingApparatus string `json:"cooking_apparatus"`
}
type RestaurantData struct {
	Name      string  `json:"name"`
	MenuItems int     `json:"menu_items"`
	Menu      []Food  `json:"menu"`
	Rating    float64 `json:"rating"`
}

type Restaurants struct {
	Info  []RegisterPayload
	Mutex sync.Mutex
}

type RegisterPayload struct {
	RestaurantData
	RestaurantId int    `json:"restaurant_id"`
	Address      string `json:"address"`
}

type MenuGet struct {
	Restaurants     int              `json:"restaurants"`
	RestaurantsData []RestaurantData `json:"restaurants_data"`
}

type Order struct {
	RestaurantId int     `json:"restaurant_id"`
	Items        []int   `json:"items"`
	Priority     int     `json:"priority"`
	MaxWait      float64 `json:"max_wait"`
	CreatedTime  int64   `json:"created_time"`
}

type Orders struct {
	ClientId int     `json:"client_id"`
	Orders   []Order `json:"orders"`
}

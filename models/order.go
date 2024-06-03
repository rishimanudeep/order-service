package models

import (
	"time"
)

type Order struct {
	OrderID          int         `json:"order_id"`
	UserID           int         `json:"user_id"`
	RestaurantID     int         `json:"restaurant_id"`
	ItemID           int         `json:"item_id"`
	RiderID          int         `json:"rider_id,omitempty"`
	Instructions     string      `json:"instructions"`
	Status           string      `json:"status"`
	PickupLocation   string      `json:"pickup_location"`
	DeliveryLocation string      `json:"delivery_location"`
	CreatedAt        time.Time   `json:"created_at"`
	DeliveryTime     time.Time   `json:"delivery_time"`
	Items            []OrderItem `json:"items"`
}

type OrderItem struct {
	ItemID     int `json:"item_id"`
	OrderID    int `json:"order_id"`
	MenuItemID int `json:"menu_item_id"`
	Quantity   int `json:"quantity"`
}

type OrderPlacedEvent struct {
	OrderID      int    `json:"order_id"`
	RestaurantID int    `json:"restaurant_id"`
	MenuID       int    `json:"menu_id"`
	Status       string `json:"status"`
}

type OrderStatusUpdate struct {
	OrderID      int     `json:"order_id"`
	RestaurantID int     `json:"restaurant_id"`
	Status       string  `json:"status"`
	ItemID       int     `json:"item_id"`
	Lat          float64 `json:"lat"`
	Long         float64 `json:"long"`
}

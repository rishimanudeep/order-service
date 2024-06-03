package models

type RiderLocation struct {
	RiderID int `json:"rider_id"`
	Location
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Availability struct {
	IsAvailable bool `json:"is_available"`
}

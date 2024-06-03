package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/errors"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"order-service/middleware"
	"order-service/models"
)

type handler struct {
	orderService OrderService
}

// New unexported return type for exported function
// used factory for injecting dependencies
func New(orderService OrderService) handler {
	return handler{orderService: orderService}
}

// CreateOrder will take the request and response and will call the service layer
// takes delivery location and userID from the middleware
func (h handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	latitude := r.Context().Value("latitude").(float64)
	longitude := r.Context().Value("longitude").(float64)

	var order models.Order

	order.UserID = userID
	order.DeliveryLocation = fmt.Sprintf("%f, %f", latitude, longitude)

	// Create a context with a deadline or timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.orderService.CreateOrder(ctx, &order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)

}

// GetOrder will takes request and response calls to service layer
func (h *handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["orderID"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrder(context.Background(), orderID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	json.NewEncoder(w).Encode(order)
}

// UpdateOrder will take request and response calls to service layer
// takes an userID from the middleware
func (h handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	latitude := r.Context().Value("latitude").(float64)
	longitude := r.Context().Value("longitude").(float64)

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["orderID"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order.OrderID = orderID
	order.UserID = userID
	order.DeliveryLocation = fmt.Sprintf("%f, %f", latitude, longitude)

	if err := h.orderService.UpdateOrder(context.Background(), &order); err != nil {
		h.handleError(w, err)
		return
	}

	json.NewEncoder(w).Encode(order)
}

// ListOrders will take request and response calls to service layer
func (h *handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var userID, riderID, restaurantID *int

	if userIDStr := queryParams.Get("user_id"); userIDStr != "" {
		id, err := strconv.Atoi(userIDStr)
		if err == nil {
			userID = &id
		}
	}

	if riderIDStr := queryParams.Get("rider_id"); riderIDStr != "" {
		id, err := strconv.Atoi(riderIDStr)
		if err == nil {
			riderID = &id
		}
	}

	if restaurantIDStr := queryParams.Get("restaurant_id"); restaurantIDStr != "" {
		id, err := strconv.Atoi(restaurantIDStr)
		if err == nil {
			restaurantID = &id
		}
	}

	orders, err := h.orderService.ListOrders(context.Background(), userID, riderID, restaurantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

// AssignRiderToOrder will take request and response calls to service layer
func (h *handler) AssignRiderToOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["orderID"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	riderID, err := strconv.Atoi(vars["riderID"])
	if err != nil {
		http.Error(w, "Invalid rider ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.AssignRiderToOrder(context.Background(), orderID, riderID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateOrderStatus will take request and response calls to service layer
func (h *handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["orderID"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var status struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.orderService.UpdateOrderStatus(context.Background(), orderID, status.Status); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) handleError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *errors.EntityNotFound:
		http.Error(w, e.Error(), http.StatusNotFound)
	case *errors.NoResponse:
		http.Error(w, e.Error(), http.StatusNotFound)
	case *errors.MissingParam:
		http.Error(w, e.Error(), http.StatusBadRequest)
	case *errors.ValidationError:
		http.Error(w, e.Error(), http.StatusBadRequest)
	case *errors.InternalServerError:
		http.Error(w, e.Error(), http.StatusInternalServerError)
	default:
		http.Error(w, "unknown error", http.StatusInternalServerError)
	}
}

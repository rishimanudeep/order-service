package handler

import (
	"context"

	"order-service/models"
)

// OrderService is an interface which have service methods in it
type OrderService interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, orderID int) (*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context, userID, riderID, restaurantID *int) ([]*models.Order, error)
	AssignRiderToOrder(ctx context.Context, orderID int, riderID int) error
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	ProcessOrderPlacedEvent(message []byte) error
	ProcessOrderStatusUpdatedEvent(msg []byte) error
}

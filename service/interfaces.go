package service

import (
	"context"
	"order-service/models"
)

type OrderStore interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, orderID int) (*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context, userID, riderID, restaurantID *int) ([]*models.Order, error)
	AssignRiderToOrder(ctx context.Context, orderID int, riderID int) error
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	UpdateRiderID(orderID int, riderID int) error
}

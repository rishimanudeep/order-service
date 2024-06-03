package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"order-service/errors"
	"order-service/models"
)

type store struct {
	db *sql.DB
}

// New unexported return type for exported function
// used for injecting db dependency
func New(db *sql.DB) store {
	return store{db: db}
}

// CreateOrder will place a new order in the order table and,
// the subsidiary items in the same order will place into the ordeItems table
func (s *store) CreateOrder(ctx context.Context, order *models.Order) error {
	ctx = context.WithValue(ctx, "db", s.db)

	// using transactions to maintain data integrity and consistency
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `INSERT INTO orders (user_id, item_id,restaurant_id, instructions, status, pickup_location, delivery_location,` +
		`created_at, delivery_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8,$9) RETURNING order_id`

	err = tx.QueryRowContext(ctx, query, order.UserID, &order.ItemID, order.RestaurantID, order.Instructions, order.Status, order.PickupLocation, order.DeliveryLocation, order.CreatedAt, order.DeliveryTime).Scan(&order.OrderID)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}

		return &errors.InternalServerError{Message: "query execution error"}
	}

	for _, item := range order.Items {
		query = `INSERT INTO order_items (order_id, menu_item_id, quantity) VALUES ($1, $2, $3)`
		_, err := tx.ExecContext(ctx, query, order.OrderID, item.MenuItemID, item.Quantity)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return &errors.InternalServerError{Message: err.Error()}
			}

			return &errors.InternalServerError{Message: "query execution error"}
		}
	}

	return tx.Commit()
}

// GetOrder will give the order details and orderItems details for specific order
func (s *store) GetOrder(ctx context.Context, orderID int) (*models.Order, error) {
	order := &models.Order{}
	query := `SELECT order_id, user_id, restaurant_id, rider_id, instructions, status, pickup_location,` +
		`delivery_location, created_at, delivery_time  FROM orders WHERE order_id = $1`

	row := s.db.QueryRowContext(ctx, query, orderID)
	if err := row.Scan(&order.OrderID, &order.UserID, &order.RestaurantID, &order.RiderID, &order.Instructions,
		&order.Status, &order.PickupLocation, &order.DeliveryLocation, &order.CreatedAt, &order.DeliveryTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.NoResponse{Message: "orders"}
		}
		return nil, &errors.InternalServerError{Message: "query execution error"}
	}

	itemsQuery := `SELECT item_id, order_id, menu_item_id, quantity FROM order_items WHERE order_id = $1`
	rows, err := s.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ItemID, &item.OrderID, &item.MenuItemID, &item.Quantity); err != nil {
			return nil, &errors.InternalServerError{Message: "scan error"}
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

// UpdateOrder will update order by orderID
func (s *store) UpdateOrder(ctx context.Context, order *models.Order) error {
	query := `UPDATE orders SET user_id = $1, restaurant_id = $2, rider_id = $3, instructions = $4, status = $5,` +
		` pickup_location = $6, delivery_location = $7, delivery_time = $8 WHERE order_id = $9`

	_, err := s.db.ExecContext(ctx, query, order.UserID, order.RestaurantID, order.RiderID, order.Instructions, order.Status,
		order.PickupLocation, order.DeliveryLocation, order.DeliveryTime, order.OrderID)
	if err != nil {
		return &errors.InternalServerError{Message: "query execution error"}
	}

	return err
}

// ListOrders will build a dynamic query where it can filter out the records based on userID,riderID, and restaurantID
func (s *store) ListOrders(ctx context.Context, userID, riderID, restaurantID *int) ([]*models.Order, error) {
	args, query := buildGetQuery(userID, riderID, restaurantID)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, &errors.InternalServerError{Message: "query execution error"}
	}
	defer rows.Close()

	var orders []*models.Order

	for rows.Next() {
		order := &models.Order{}
		// to handle empty values
		var rider_id sql.NullInt64
		var instructions sql.NullString
		var deliveryTime sql.NullTime

		if err := rows.Scan(&order.OrderID, &order.UserID, &order.RestaurantID, &rider_id, &instructions,
			&order.Status, &order.PickupLocation, &order.DeliveryLocation, &order.CreatedAt, &deliveryTime); err != nil {
			return nil, err
		}

		if rider_id.Valid {
			order.RiderID = int(rider_id.Int64)
		}

		if instructions.Valid {
			order.Instructions = instructions.String
		}

		if deliveryTime.Valid {
			order.DeliveryTime = deliveryTime.Time
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, &errors.InternalServerError{Message: "rows error"}
	}

	return orders, nil
}

// AssignRiderToOrder will set rider for the order
func (s *store) AssignRiderToOrder(ctx context.Context, orderID int, riderID int) error {
	query := `UPDATE orders SET rider_id = $1 WHERE order_id = $2`

	_, err := s.db.ExecContext(ctx, query, riderID, orderID)
	if err != nil {
		return &errors.InternalServerError{Message: "Query Execution Error"}
	}
	return nil
}

// UpdateOrderStatus will update the status for the order
func (s *store) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE order_id = $2`

	_, err := s.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return &errors.InternalServerError{Message: "Query Execution Error"}
	}

	return nil
}

// UpdateRiderID updates the rider ID for a given order ID
func (s *store) UpdateRiderID(orderID int, riderID int) error {
	query := "UPDATE orders SET rider_id = $1 WHERE order_id = $2"

	_, err := s.db.Exec(query, riderID, orderID)
	if err != nil {
		return &errors.InternalServerError{Message: "query execution failed"}
	}

	return nil
}

// support function for building getAllQuery
func buildGetQuery(userID *int, riderID *int, restaurantID *int) ([]interface{}, string) {
	var conditions []string
	var args []interface{}
	argPosition := 1

	// checks for non-empty cases
	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPosition))
		args = append(args, *userID)
		argPosition++
	}
	if riderID != nil {
		conditions = append(conditions, fmt.Sprintf("rider_id = $%d", argPosition))
		args = append(args, *riderID)
		argPosition++
	}
	if restaurantID != nil {
		conditions = append(conditions, fmt.Sprintf("restaurant_id = $%d", argPosition))
		args = append(args, *restaurantID)
		argPosition++
	}

	query := `SELECT order_id, user_id, restaurant_id, rider_id, instructions, status, pickup_location, delivery_location,` +
		`created_at, delivery_time FROM orders`

	// build where condition based on the conditions
	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}
	return args, query
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/IBM/sarama"

	"order-service/errors"
	"order-service/models"
)

// service struct will be used send as receiver
type service struct {
	orderStore   OrderStore
	kProduce     sarama.SyncProducer
	riderService *http.Client
}

// New factory function will have injected dependencies
func New(orderStore OrderStore, producer sarama.SyncProducer, riderService *http.Client) service {
	return service{orderStore: orderStore, kProduce: producer, riderService: riderService}
}

// CreateOrder will create an order status as pending and updates restaurant service of order status via event driven
func (s *service) CreateOrder(ctx context.Context, order *models.Order) error {
	// check mandatory fields of order
	err := validate(order)
	order.CreatedAt = time.Now()
	// default status as pending
	order.Status = "pending"

	// calls to store for creation of order
	err = s.orderStore.CreateOrder(ctx, order)
	if err != nil {
		return err
	}

	// will send an event to restaurant of placed order
	if err := s.publishOrderPlacedEvent(order); err != nil {
		return err
	}

	return nil
}

// publishOrderPlacedEvent is a helper method which produces an event in kafka topic
func (s *service) publishOrderPlacedEvent(order *models.Order) error {
	// event struct
	event := models.OrderPlacedEvent{
		OrderID:      order.OrderID,
		RestaurantID: order.RestaurantID,
		MenuID:       order.ItemID,
		Status:       order.Status,
	}

	// marshalling the event to sendMessage
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// kafkaMessage will consists of Topic and value
	kafkaMessage := &sarama.ProducerMessage{
		Topic: "order-placed",
		Value: sarama.StringEncoder(eventJSON),
	}

	// sends message to kafka topic
	partition, offset, err := s.kProduce.SendMessage(kafkaMessage)
	if err != nil {
		return &errors.InternalServerError{Message: "kafka producer error"}
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

// GetOrder calls to store layer for to fetch orderID details
func (s *service) GetOrder(ctx context.Context, orderID int) (*models.Order, error) {
	return s.orderStore.GetOrder(ctx, orderID)
}

// UpdateOrder calls to store layer for to update orderID details
func (s *service) UpdateOrder(ctx context.Context, order *models.Order) error {
	return s.orderStore.UpdateOrder(ctx, order)
}

// ListOrders calls to store layer for to fetch orders details, filtering orders by parameters
func (s *service) ListOrders(ctx context.Context, userID, riderID, restaurantID *int) ([]*models.Order, error) {
	return s.orderStore.ListOrders(ctx, userID, riderID, restaurantID)
}

// AssignRiderToOrder calls to store layer for to assign a rider for the order
func (s *service) AssignRiderToOrder(ctx context.Context, orderID int, riderID int) error {
	return s.orderStore.AssignRiderToOrder(ctx, orderID, riderID)
}

// UpdateOrderStatus calls to store layer to update the status of order
// And will call rider service to update the rider status
func (s *service) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	err := s.orderStore.UpdateOrderStatus(ctx, orderID, status)
	if err != nil {
		return err
	}

	if status == "delivered" {
		order, err := s.orderStore.GetOrder(ctx, orderID)
		if err != nil {
			return err
		}

		var riderAvailable models.Availability

		riderAvailable.IsAvailable = true
		err = s.updateRiderAvailabilty(riderAvailable, order.RiderID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ProcessOrderStatusUpdatedEvent once the restaurant service updated, if the status is accepted it assigns rider
// by calling rider service
func (s *service) ProcessOrderStatusUpdatedEvent(msg []byte) error {
	var event models.OrderStatusUpdate
	// unmarshall msg to event
	err := json.Unmarshal(msg, &event)
	if err != nil {
		log.Printf("Failed to unmarshal Kafka message: %v", err)
		return err
	}

	// initialize context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// update the status to the order
	err = s.orderStore.UpdateOrderStatus(ctx, event.OrderID, event.Status)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		return err
	}

	// if status is accepted it gets available rider and update the riderID in orders table
	if event.Status == "accepted" {
		riders, err := s.getAvailableRiders(event)
		if err != nil {
			return err
		}
		if len(riders) == 0 {
			return &errors.EntityNotFound{"No riders available near restaurant"}
		}

		var riderAvailable models.Availability
		riderAvailable.IsAvailable = false
		s.orderStore.UpdateRiderID(event.OrderID, riders[0].RiderID)

		err = s.updateRiderAvailabilty(riderAvailable, riders[0].RiderID)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Order status updated: %d -> %s\n", event.OrderID, event.Status)

	return nil
}

// getAvailableRiders is a helper method which calls to rider service to fetch nearby riders
func (s *service) getAvailableRiders(event models.OrderStatusUpdate) ([]models.RiderLocation, error) {
	// Use floating-point formatting for latitude and longitude
	url := fmt.Sprintf("http://localhost:8080/riders/nearby?latitude=%f&longitude=%f&radius=%d", event.Lat, event.Long, 3)
	log.Println(url)
	// Remove spaces from the URL
	url = strings.TrimSpace(url)

	resp, err := s.riderService.Get(url)
	if err != nil {
		return nil, &errors.InternalServerError{"failed to make HTTP request to Rider Service"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &errors.InternalServerError{Message: "rider Service returned non-OK status code"}
	}

	var rider []models.RiderLocation
	if err := json.NewDecoder(resp.Body).Decode(&rider); err != nil {
		return nil, &errors.InternalServerError{"failed to decode response body"}
	}

	return rider, nil
}

// updateRiderAvailability will update status of rider via calling the rider service
func (s *service) updateRiderAvailabilty(availability models.Availability, id int) error {
	url := fmt.Sprintf("http://localhost:8080/rider/%d/availability", id)

	jsonData, err := json.Marshal(availability)
	if err != nil {
		return fmt.Errorf("failed to marshal availability details: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &errors.InternalServerError{Message: "failed to create request"}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.riderService.Do(req)
	if err != nil {
		return &errors.InternalServerError{Message: "failed to make request"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body for additional error details
		_, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return &errors.InternalServerError{"unexpected status code and failed to read response body"}
		}
		return &errors.InternalServerError{"unexpected status code response"}
	}

	return nil
}

// validate checks for the mandatory fields
func validate(o *models.Order) error {
	if o.RestaurantID == 0 {
		return &errors.MissingParam{Message: "Restaurant ID"}
	}
	if o.ItemID == 0 {
		return &errors.MissingParam{Message: "item ID is required"}
	}
	if o.Instructions == "" {
		return &errors.MissingParam{Message: "instructions are required"}
	}
	if o.PickupLocation == "" {
		return &errors.MissingParam{Message: "pickup location is required"}
	}
	return nil
}

func (s *service) ProcessOrderPlacedEvent(message []byte) error {
	var event models.OrderPlacedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	// Process the event and update the database
	fmt.Printf("Processing order placed event: %+v\n", event)
	// Implement the logic to update the order status in the database
	ctx, cancel := context.WithTimeout(nil, 60*time.Second)
	defer cancel()

	err := s.UpdateOrderStatus(ctx, event.OrderID, event.Status)
	if err != nil {
		return err
	}

	return nil
}

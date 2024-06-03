package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"gofr.dev/pkg/gofr"

	"order-service/handler"
	"order-service/middleware"
	"order-service/migrations"
	"order-service/service"
	"order-service/store"
)

func main() {
	// connection string
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		"postgres",  // DB_USER
		"root123",   // DB_PASSWORD
		"user_db",   // DB_NAME
		"localhost", // DB_HOST
		"5432",      // DB_PORT
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// gofr framework using for migrations
	a := gofr.New()
	a.Migrate(migrations.All())

	riderService := &http.Client{}

	// Initialize Kafka producer
	kafkaProducer, err := NewKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// injecting dependencies to each layer
	oderStore := store.New(db)
	orderService := service.New(&oderStore, kafkaProducer, riderService)
	orderHandler := handler.New(&orderService)

	r := mux.NewRouter()

	r.Use(middleware.JWTMiddleware) // Apply JWT middleware

	r.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/orders/{orderID}", orderHandler.GetOrder).Methods("GET")
	r.HandleFunc("/orders/{orderID}", orderHandler.UpdateOrder).Methods("PUT")
	r.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")
	r.HandleFunc("/orders/{orderID}/assign/{riderID}", orderHandler.AssignRiderToOrder).Methods("POST")
	r.HandleFunc("/orders/{orderID}/status", orderHandler.UpdateOrderStatus).Methods("PUT")

	// Start HTTP server
	go func() {
		log.Println("Starting server on port 8000")
		if err := http.ListenAndServe(":8000", r); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start Kafka consumer to listen for messages from the "order-placed" topic
	go func() {
		if err := startKafkaConsumer(&orderService); err != nil {
			log.Fatalf("Failed to start Kafka consumer: %v", err)
		}
	}()

	// Waits for termination signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh

	log.Println("Shutting down server...")
}

// NewKafkaProducer will connect to kafka port and initializes the producer
func NewKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kafka producer: %v", err)
	}

	return producer, nil
}

// startKafkaConsumer will check for update events on topic
func startKafkaConsumer(orderHandler handler.OrderService) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		return fmt.Errorf("error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("order-status-updated", 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("error subscribing to Kafka topic: %v", err)
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Received message: %s\n", string(msg.Value))
			// Implement logic to process Kafka message and update order database
			err := orderHandler.ProcessOrderStatusUpdatedEvent(msg.Value)
			if err != nil {
				return err
			}
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}

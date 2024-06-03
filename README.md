# Order-Service

This is a Go application for managing order service.  The Order Service is responsible for managing the creation, updating, retrieval, and status management of customer orders, ensuring data consistency and integrity through transactional operations and integrating with external services for order fulfillment.

## Features

- Order Management: Create, update, and retrieve customer orders.
- Rider Integration: Calls the Rider Service to find nearby riders for order delivery.
- Event-Driven Updates: Uses event-driven architecture to update the status of orders and notify the Restaurant Service about order status changes.
- Transactional Integrity: Utilizes database transactions to maintain data consistency and integrity.

## Getting Started with Order-Service

### Requirements

- A working Go environment - [https://go.dev/dl/](https://go.dev/dl/)
- Check the go version with command: go version.
- One should also be familiar with the Golang syntax. [Golang Tour](https://tour.golang.org/) has an excellent guided tour and highly recommended.

### Installation

## GOFR as dependency used for migrations

- To get the GOFR as a dependency, use the command:
  `go get gofr.dev`

- Then use the command `go mod tidy`, to download the necessary packages.


### To Run Server

Run `go run main.go` command in CLI.

## Usage

The application provides the following RESTful endpoints:

- `POST /orders`: Place a new order.
- `PUT /orders/{orderID}`: Update a new order by ID.
- `GET /orders/{orderID}"`: Retrieve an Order by ID.
- `GET /orders`: Fetch all orders (filter enabled).
- `POST /orders/{orderID}/assign/{riderID}`: Rider assign to an Order.
- `PUT "/orders/{orderID}/status"`: Updates Order Status.

## Dependencies

The application uses the following dependencies:

- `gofr.dev/pkg/gofr`: A Go web framework used for handling HTTP requests.
- `Order-service/handlers`: Handlers package for handling HTTP requests related to orders.
- `Order-service/services`: Services package for business logic related to orders.
- `Order-service/stores`:Store package for handling db operations realated to orders

For any information please reach out to me via rishimanudeepg@gmail.com
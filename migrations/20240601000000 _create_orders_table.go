package migrations

import (
	"gofr.dev/pkg/gofr/migration"
	"log"
)

const createOrderTable = `CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    item_id INT NOT NULL,
    restaurant_id INT NOT NULL,
    rider_id INT,
    instructions TEXT,
    status VARCHAR(50) NOT NULL,
    pickup_location POINT,
    delivery_location POINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    delivery_time TIMESTAMPTZ,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);`

func orderTable() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			_, err := d.SQL.Exec(createOrderTable)
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		},
	}
}

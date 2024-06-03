package migrations

import "gofr.dev/pkg/gofr/migration"

const order_items = `CREATE TABLE order_items (
                             item_id SERIAL PRIMARY KEY,
                             order_id INT NOT NULL,
                             menu_item_id INT NOT NULL,
                             quantity INT NOT NULL,
                             CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(order_id)
);`

func orderItemsTable() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			_, err := d.SQL.Exec(order_items)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

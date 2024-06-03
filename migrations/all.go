package migrations

import "gofr.dev/pkg/gofr/migration"

func All() map[int64]migration.Migrate {
	return map[int64]migration.Migrate{
		20240601000000: orderTable(),
		20240601153000: orderItemsTable(),
	}
}

package migration

import (
	paymentEntity "vibe-ddd-golang/internal/application/payment/entity"
	userEntity "vibe-ddd-golang/internal/application/user/entity"
	database "vibe-ddd-golang/internal/pkg/db"

	"ariga.io/atlas-provider-gorm/gormschema"
)

// entities is the single source of truth for the GORM models this service owns. Atlas
// migrations remain authoritative; this list drives dev DB_SYNC auto-migrate and
// `make migrate-diff`.
func entities() []any {
	return []any{
		userEntity.User{},
		paymentEntity.Payment{},
	}
}

func MigrationEntity(db *database.Database) error {
	return db.Migrate(entities()...)
}

func ListTableEntity() []string {
	return []string{
		userEntity.User{}.TableName(),
		paymentEntity.Payment{}.TableName(),
	}
}

func MigrationServer(driver database.DriverEnum) (*string, error) {
	stmts, err := gormschema.New(driver.ToString()).Load(entities()...)

	return &stmts, err
}

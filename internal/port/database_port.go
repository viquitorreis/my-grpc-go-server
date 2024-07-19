package port

import (
	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-go-server/internal/adapter/database"
)

type DummyDatabasePort interface {
	Save(data *database.DummyOrm) (uuid.UUID, error)
	GetByUUID(uuid uuid.UUID) (database.DummyOrm, error)
}

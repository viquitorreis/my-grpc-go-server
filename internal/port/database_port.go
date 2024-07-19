package port

import (
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-go-server/internal/adapter/database"
)

type DummyDatabasePort interface {
	Save(data *database.DummyOrm) (uuid.UUID, error)
	GetByUUID(uuid uuid.UUID) (database.DummyOrm, error)
}

type BankDatabasePort interface {
	GetBankAccountNumber(account string) (database.BankAccountOrm, error)
	CreateExchangeRate(r database.BankExchangeRateOrm) (uuid.UUID, error)
	GetExchangeRate(fromCurrency, toCurrency string, ts time.Time) (database.BankExchangeRateOrm, error)
}

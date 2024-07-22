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
	CreateTransaction(account database.BankAccountOrm, t database.BankTransactionOrm) (uuid.UUID, error)
	CreateTransfer(transfer database.BankTransferOrm) (uuid.UUID, error)
	CreateTransferTransactionPair(fromAccountOrm database.BankAccountOrm, toAccountOrm database.BankAccountOrm,
		fromTransactionOrm database.BankTransactionOrm, toTransactionOrm database.BankTransactionOrm) (bool, error)
	UpdateTransferStatus(transfer database.BankTransferOrm, status bool) error
}

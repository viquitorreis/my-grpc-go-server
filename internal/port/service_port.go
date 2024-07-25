package port

import (
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
)

type HelloServicePort interface {
	GenerateHello(name string) string
}

type BankServicePort interface {
	FindCurrentBalance(accountId string) (float64, error)
	CreateExchangeRate(r bank.ExchangeRate) (uuid.UUID, error)
	GetExchangeRate(fromCurrency, toCurrency string, ts time.Time) (float64, error)
	CreateTransaction(account string, t bank.Transaction) (uuid.UUID, error)
	CalculateTransactionSummary(tsum *bank.TransactionSummary, trans bank.Transaction) error
	Transfer(tt bank.TransferTransaction) (uuid.UUID, bool, error)
}

type ResiliencyServicePort interface {
	GenerateResiliency(minDelaySec int32, maxDelaySec int32, statusCodes []uint32) (string, uint32)
}

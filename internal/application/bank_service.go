package application

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-go-server/internal/adapter/database"
	"github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
	"github.com/viquitorreis/my-grpc-go-server/internal/port"
)

type BankService struct {
	db port.BankDatabasePort
}

func NewBankService(port port.BankDatabasePort) *BankService {
	return &BankService{db: port}
}

func (s *BankService) FindCurrentBalance(accountId string) float64 {
	bankAccount, err := s.db.GetBankAccountNumber(accountId)
	if err != nil {
		log.Printf("failed to get bank account number: %v\n", err)
		return 0.0
	}

	return bankAccount.CurrentBalance
}

func (s *BankService) CreateExchangeRate(r bank.ExchangeRate) (uuid.UUID, error) {
	newUUID := uuid.New()
	now := time.Now()

	exchangeRateOrm := database.BankExchangeRateOrm{
		ExchangeRateUUID:   newUUID,
		FromCurrency:       r.FromCurrency,
		ToCurrency:         r.ToCurrency,
		Rate:               r.Rate,
		ValidFromTimestamp: r.ValidFromTimestamp,
		ValidToTimestamp:   r.ValidToTimestamp,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	return s.db.CreateExchangeRate(exchangeRateOrm)
}

func (s *BankService) GetExchangeRate(fromCurrency, toCurrency string, ts time.Time) float64 {
	exchangeRate, err := s.db.GetExchangeRate(fromCurrency, toCurrency, ts)
	if err != nil {
		return 0
	}

	return float64(exchangeRate.Rate)
}

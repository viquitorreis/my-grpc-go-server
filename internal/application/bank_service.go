package application

import (
	"fmt"
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

func (s *BankService) CreateTransaction(account string, t bank.Transaction) (uuid.UUID, error) {
	newUUID := uuid.New()
	now := time.Now()

	bankAccOrm, err := s.db.GetBankAccountNumber(account)
	if err != nil {
		log.Printf("failed to get bank account number: %v\n", err)
		return uuid.Nil, err
	}

	transactionOrm := database.BankTransactionOrm{
		TransactionUUID:      newUUID,
		AccountUUID:          bankAccOrm.AccountUUID,
		TransactionTimestamp: now,
		Amount:               t.Amount,
		TransactionType:      t.TransactionType,
		Notes:                t.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	savedUUID, err := s.db.CreateTransaction(bankAccOrm, transactionOrm)
	return savedUUID, err
}

func (s *BankService) CalculateTransactionSummary(tsum *bank.TransactionSummary, trans bank.Transaction) error {
	switch trans.TransactionType {
	case bank.TransactionTypeIn:
		tsum.SumIn += trans.Amount
	case bank.TransactionTypeOut:
		tsum.SumOut += trans.Amount
	default:
		return fmt.Errorf("unknown transaction type %v", trans.TransactionType)
	}

	tsum.SumTotal = tsum.SumIn - tsum.SumOut

	return nil
}

func (s *BankService) Transfer(tt bank.TransferTransaction) (uuid.UUID, bool, error) {
	now := time.Now()

	fromAccOrm, err := s.db.GetBankAccountNumber(tt.FromAccountNumber)
	if err != nil {
		log.Printf("failed to get bank account number: %v\n", err)
		return uuid.Nil, false, err
	}

	toAccOrm, err := s.db.GetBankAccountNumber(tt.ToAccountNumber)
	if err != nil {
		log.Printf("failed to get bank account number: %v\n", err)
		return uuid.Nil, false, err
	}

	fromTransactionOrm := database.BankTransactionOrm{
		TransactionUUID:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      bank.TransactionTypeOut,
		AccountUUID:          fromAccOrm.AccountUUID,
		Amount:               tt.Amount,
		Notes:                "Transfer to " + tt.ToAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	toTransactionOrm := database.BankTransactionOrm{
		TransactionUUID:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      bank.TransactionTypeIn,
		AccountUUID:          toAccOrm.AccountUUID,
		Amount:               tt.Amount,
		Notes:                "Transfer from " + tt.FromAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// create transfer request
	newTransferUUID := uuid.New()

	transferOrm := database.BankTransferOrm{
		TransferUUID:      newTransferUUID,
		FromAccountUUID:   fromAccOrm.AccountUUID,
		Currency:          tt.Currency,
		Amount:            tt.Amount,
		TransferTimestamp: now,
		TransferSuccess:   false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if _, err := s.db.CreateTransfer(transferOrm); err != nil {
		log.Printf("failed to create transfer de %v para %v : %v\n", tt.FromAccountNumber, tt.ToAccountNumber, err)
		return uuid.Nil, false, err
	}

	if transferPairsucess, err := s.db.CreateTransferTransactionPair(fromAccOrm, toAccOrm, fromTransactionOrm, toTransactionOrm); transferPairsucess {
		s.db.UpdateTransferStatus(transferOrm, true)
		return newTransferUUID, true, nil
	} else {
		return newTransferUUID, false, err
	}
}

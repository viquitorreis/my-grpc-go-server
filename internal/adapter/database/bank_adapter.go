package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func (a *DatabaseAdapter) GetBankAccountNumber(account string) (BankAccountOrm, error) {
	var bankAccountOrm BankAccountOrm
	if err := a.db.First(&bankAccountOrm, "account_number = ?", account).Error; err != nil {
		log.Printf("failed to get bank account number: %v\n", err)
		return bankAccountOrm, fmt.Errorf("failed to get bank account number: %w", err)
	}

	return bankAccountOrm, nil
}

func (a *DatabaseAdapter) CreateExchangeRate(r BankExchangeRateOrm) (uuid.UUID, error) {
	if err := a.db.Create(&r).Error; err != nil {
		log.Printf("failed to create exchange rate: %v\n", err)
		return uuid.Nil, fmt.Errorf("failed to create exchange rate: %w", err)
	}

	return r.ExchangeRateUUID, nil
}

func (a *DatabaseAdapter) GetExchangeRate(fromCurrency, toCurrency string, ts time.Time) (BankExchangeRateOrm, error) {
	var exchangeRateOrm BankExchangeRateOrm

	err := a.db.First(&exchangeRateOrm, `
		from_currency = ?
		AND to_currency = ? 
		AND (? BETWEEN valid_from_timestamp AND valid_to_timestamp)
	`, fromCurrency, toCurrency, ts).Error

	return exchangeRateOrm, err
}

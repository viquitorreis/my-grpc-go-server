package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
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

func (a *DatabaseAdapter) CreateTransaction(account BankAccountOrm, t BankTransactionOrm) (uuid.UUID, error) {
	tx := a.db.Begin() // começa a transação

	if err := tx.Create(&t).Error; err != nil {
		tx.Rollback() // rollback da transação caso ocorra erro
		return uuid.Nil, err
	}

	// recalcula o saldo da conta
	newAmount := t.Amount

	if t.TransactionType == bank.TransactionTypeOut {
		newAmount = -1 * t.Amount
	}

	newAccountBalance := account.CurrentBalance + newAmount

	// atualizando o saldo da conta e o updated_at
	if err := tx.Model(&account).Updates(
		map[string]interface{}{
			"current_balance": newAccountBalance,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	tx.Commit() // commit da transação

	return t.AccountUUID, nil
}

func (a *DatabaseAdapter) CreateTransfer(transfer BankTransferOrm) (uuid.UUID, error) {
	if err := a.db.Create(&transfer).Error; err != nil {
		return uuid.Nil, err
	}

	return transfer.TransferUUID, nil
}

func (a *DatabaseAdapter) CreateTransferTransactionPair(fromAccountOrm BankAccountOrm, toAccountOrm BankAccountOrm,
	fromTransactionOrm BankTransactionOrm, toTransactionOrm BankTransactionOrm) (bool, error) {
	tx := a.db.Begin()

	if err := tx.Create(&fromTransactionOrm).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Create(&toTransactionOrm).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// recalculando o balanço da conta de origem
	fromAccNewBal := fromAccountOrm.CurrentBalance - fromTransactionOrm.Amount

	if err := tx.Model(&fromAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": fromAccNewBal,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// recalculando o balanço da conta de destino
	toAccNewBal := toAccountOrm.CurrentBalance + toTransactionOrm.Amount

	if err := tx.Model(&toAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": toAccNewBal,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()

	return true, nil
}

func (a *DatabaseAdapter) UpdateTransferStatus(transfer BankTransferOrm, status bool) error {
	if err := a.db.Model(&transfer).Updates(
		map[string]interface{}{
			"transfer_success": status,
			"updated_at":       time.Now(),
		},
	).Error; err != nil {
		return err
	}

	return nil
}

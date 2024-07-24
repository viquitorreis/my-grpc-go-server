package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	domainBank "github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank.CurrentBalanceRequest) (*bank.CurrentBalanceResponse, error) {
	now := time.Now()
	bal, err := a.bankService.FindCurrentBalance(req.AccountNumber)
	if err != nil {
		log.Printf("failed to get current balance: %v\n", err)
		return nil, status.Error(codes.FailedPrecondition, "failed to get current balance")
	}

	return &bank.CurrentBalanceResponse{
		Amount: bal,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}

func (a *GrpcAdapter) FetchExchangeRates(req *bank.ExchangeRateRequest, stream bank.BankService_FetchExchangeRatesServer) error {
	context := stream.Context()

	for {
		select {
		case <-context.Done():
			log.Println("client cancelou o streaming")
			return nil
		default:
			now := time.Now().Truncate(time.Second)
			rate, err := a.bankService.GetExchangeRate(req.FromCurrency, req.ToCurrency, now)
			if err != nil {
				log.Printf("failed to get exchange rate: %v\n", err)
				s := status.New(codes.FailedPrecondition, "failed to get exchange rate")
				s, _ = s.WithDetails(&errdetails.ErrorInfo{
					Domain: "bank.com",
					Reason: "failed to get exchange rate",
					Metadata: map[string]string{
						"from_currency": req.FromCurrency,
						"to_currency":   req.ToCurrency,
					},
				})

				return s.Err()
			}

			stream.Send(&bank.ExchangeRateResponse{
				FromCurrency: req.FromCurrency,
				ToCurrency:   req.ToCurrency,
				Rate:         rate,
				Timestamp:    now.Format(time.RFC3339),
			})

			log.Printf("Exchange rate send to client, %v to %v: %v\n", req.FromCurrency, req.ToCurrency, rate)

			time.Sleep(3 * time.Second)
		}
	}
}

func (a *GrpcAdapter) SummarizeTransactions(stream bank.BankService_SummarizeTransactionsServer) error {
	tsum := domainBank.TransactionSummary{
		SummaryOnDate: time.Now(),
		SumIn:         0,
		SumOut:        0,
		SumTotal:      0,
	}

	account := ""

	// loop infinito para receber as conexões do client
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			res := bank.TransactionSummary{
				AccountNumber: account,
				SumAmountIn:   tsum.SumIn,
				SumAmountOut:  tsum.SumOut,
				SumTotal:      tsum.SumTotal,
				TransactionDate: &date.Date{
					Year:  int32(tsum.SummaryOnDate.Year()),
					Month: int32(tsum.SummaryOnDate.Month()),
					Day:   int32(tsum.SummaryOnDate.Day()),
				},
			}

			return stream.SendAndClose(&res)
		}

		if err != nil {
			log.Fatalf("failed to receive transaction from client: %v\n", err)
		}

		ts, err := toTime(req.Timestamp)
		if err != nil {
			log.Fatalf("failed to convert timestamp: %v\n", err)
		}

		tranType := domainBank.TransactionTypeUnknown

		if req.Type == bank.TransactionType_TRANSACTION_TYPE_IN {
			tranType = domainBank.TransactionTypeIn
		} else if req.Type == bank.TransactionType_TRANSACTION_TYPE_OUT {
			tranType = domainBank.TransactionTypeOut
		}

		tcurrent := domainBank.Transaction{
			Amount:          req.Amount,
			Timestamp:       ts,
			TransactionType: tranType,
		}

		accUUID, err := a.bankService.CreateTransaction(req.AccountNumber, tcurrent)
		if err != nil && accUUID == uuid.Nil {
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "account_number",
						Description: "invalid account number",
					},
				},
			})

			return s.Err()
		} else if err != nil && accUUID != uuid.Nil {
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "amount",
						Description: fmt.Sprintf("invalid amount: %v. Exceeds available balance", req.Amount),
					},
				},
			})

			return s.Err()
		}

		// chamando camada service para calular o resumo da transação
		err = a.bankService.CalculateTransactionSummary(&tsum, tcurrent)
		if err != nil {
			log.Printf("failed to calculate transaction summary: %v\n", err)
			return err
		}
	}
}

func currentDatetime() *datetime.DateTime {
	now := time.Now()

	return &datetime.DateTime{
		Year:       int32(now.Year()),
		Month:      int32(now.Month()),
		Day:        int32(now.Day()),
		Hours:      int32(now.Hour()),
		Minutes:    int32(now.Minute()),
		Seconds:    int32(now.Second()),
		Nanos:      int32(now.Second()),
		TimeOffset: &datetime.DateTime_UtcOffset{},
	}
}

func toTime(dt *datetime.DateTime) (time.Time, error) {
	if dt == nil {
		now := time.Now()

		dt = &datetime.DateTime{
			Year:    int32(now.Year()),
			Month:   int32(now.Month()),
			Day:     int32(now.Day()),
			Hours:   int32(now.Hour()),
			Minutes: int32(now.Minute()),
			Seconds: int32(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		}
	}

	res := time.Date(
		int(dt.Year), time.Month(dt.Month), int(dt.Day),
		int(dt.Hours), int(dt.Minutes), int(dt.Seconds), int(dt.Nanos), time.UTC,
	)

	return res, nil
}

func (a *GrpcAdapter) TransferMultiple(stream bank.BankService_TransferMultipleServer) error {
	context := stream.Context()

	for {
		select {
		case <-context.Done():
			log.Println("client cancelou o streaming")
			return nil
		default:
			req, err := stream.Recv()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				log.Printf("failed to receive transaction from client: %v\n", err)
			}

			tt := domainBank.TransferTransaction{
				FromAccountNumber: req.FromAccountNumber,
				ToAccountNumber:   req.ToAccountNumber,
				Currency:          req.Currency,
				Amount:            req.Amount,
			}

			_, tansferSuccess, err := a.bankService.Transfer(tt)
			if err != nil {
				log.Printf("failed to transfer transaction: %v\n", err)
				return err
			}

			res := bank.TransferResponse{
				FromAccountNumber: req.FromAccountNumber,
				ToAccountNumber:   req.ToAccountNumber,
				Currency:          req.Currency,
				Amount:            req.Amount,
				Timestamp:         currentDatetime(),
			}

			if tansferSuccess {
				res.Status = bank.TransferStatus_TRANSFER_STATUS_SUCCESS
			} else {
				res.Status = bank.TransferStatus_TRANSFER_STATUS_FAILED
			}

			err = stream.Send(&res)
			if err != nil {
				log.Printf("failed to send transfer response: %v\n", err)
				return err
			}
		}
	}
}

package grpc

import (
	"context"
	"io"
	"log"
	"time"

	domainBank "github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank.CurrentBalanceRequest) (*bank.CurrentBalanceResponse, error) {
	now := time.Now()
	bal := a.bankService.FindCurrentBalance(req.AccountNumber)
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
			rate := a.bankService.GetExchangeRate(req.FromCurrency, req.ToCurrency, now)

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
			log.Fatalln("failed to receive transaction from client: %v\n", err)
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

		_, err = a.bankService.CreateTransaction(req.AccountNumber, tcurrent)
		if err != nil {
			log.Printf("failed to create transaction: %v\n", err)
			return err
		}

		// chamando camada service para calular o resumo da transação
		err = a.bankService.CalculateTransactionSummary(&tsum, tcurrent)
		if err != nil {
			log.Printf("failed to calculate transaction summary: %v\n", err)
			return err
		}
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

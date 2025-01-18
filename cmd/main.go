package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	db "github.com/viquitorreis/my-grpc-go-server/db/migrations"
	"github.com/viquitorreis/my-grpc-go-server/internal/adapter/database"
	mygrpc "github.com/viquitorreis/my-grpc-go-server/internal/adapter/grpc"
	app "github.com/viquitorreis/my-grpc-go-server/internal/application"
	"github.com/viquitorreis/my-grpc-go-server/internal/application/domain/bank"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	// docker run --name my-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=postgres -p 5432:5432 -d postgres
	pgDB, err := sql.Open("postgres", "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	db.Migrate(pgDB)

	databaseAdapter, err := database.NewDatabaseAdapter(pgDB)
	if err != nil {
		log.Fatalf("Error creating database adapter: %v", err)
	}

	hs := &app.HelloService{}
	bs := app.NewBankService(databaseAdapter)
	rs := &app.ResiliencyService{}

	go generateExchangeRates(bs, "USD", "BRL", 5*time.Second)

	grpcAdapter := mygrpc.NewGrpcAdapter(hs, bs, rs, 9090)
	grpcAdapter.Run()
}

func runDummyOrm(da *database.DatabaseAdapter) {
	now := time.Now()
	uuid, err := da.Save(&database.DummyOrm{
		UserId:    uuid.New(),
		Name:      "Victor" + time.Now().Format("15:04:05"),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		log.Fatalf("Error saving data: %v", err)
	}

	res, err := da.GetByUUID(uuid)
	if err != nil {
		log.Fatalf("Error getting data: %v", err)
	}

	log.Println("Data saved and retrieved successfully: ", res)
}

func generateExchangeRates(bs *app.BankService, fromCurrency, toCurrency string, duration time.Duration) {
	ticker := time.NewTicker(duration)

	for range ticker.C {
		now := time.Now()
		validFrom := now.Truncate(time.Second).Add(3 * time.Second)
		validTo := validFrom.Add(-1 * time.Millisecond)

		// exchange rate a cada loop
		dummyRate := bank.ExchangeRate{
			FromCurrency:       fromCurrency,
			ToCurrency:         toCurrency,
			ValidFromTimestamp: validFrom,
			ValidToTimestamp:   validTo,
			Rate:               2000 + float64(rand.Intn(300)),
		}

		bs.CreateExchangeRate(dummyRate)
	}
}

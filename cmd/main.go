package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	db "github.com/viquitorreis/my-grpc-go-server/db/migrations"
	"github.com/viquitorreis/my-grpc-go-server/internal/adapter/database"
	mygrpc "github.com/viquitorreis/my-grpc-go-server/internal/adapter/grpc"
	app "github.com/viquitorreis/my-grpc-go-server/internal/application"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	pgDB, err := sql.Open("postgres", "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	db.Migrate(pgDB)

	databaseAdapter, err := database.NewDatabaseAdapter(pgDB)
	if err != nil {
		log.Fatalf("Error creating database adapter: %v", err)
	}
	runDummyOrm(databaseAdapter)

	hs := &app.HelloService{}
	bs := &app.BankService{}

	grpcAdapter := mygrpc.NewGrpcAdapter(hs, bs, 9090)
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

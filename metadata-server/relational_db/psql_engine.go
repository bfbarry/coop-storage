package relational_db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db/generated"
	"os"
)

type DbPoolEngine struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

var PSQL DbPoolEngine

func (e *DbPoolEngine) Start() {
	connStr := os.Getenv("DATABASE_URL")
	fmt.Printf("HELLO THE URL: %s\n", connStr)
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	e.pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	err = e.pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	e.Queries = db.New(e.pool)
	fmt.Println("Successfully connected to the database pool.")
}

func (e *DbPoolEngine) Stop() {
	e.pool.Close()
}

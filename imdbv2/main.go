package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adamplansky/go-bridge-mentoring/imdbv2/persistance"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

type Link struct {
	Uid   string   `json:"uid,omitempty"`
	URL   string   `json:"url,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type App struct {
	log *zap.SugaredLogger
}

// https://dgraph.io/docs/clients/go/
func (a App) Run(ctx context.Context) error {
	//dsn := os.Getenv("DGRAPH_DSN")
	//apiKey := os.Getenv("DGRAPH_APIKEY")
	//if dsn == "" {
	//	dsn = "localhost:9080"
	//}
	const apiKey = ""
	const dsn = "localhost:9080"
	db, err := persistance.NewDB(dsn, apiKey)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer db.Close()

	if err = db.Migrate(ctx, true); err != nil {
		return fmt.Errorf("db migrate failed: %w", err)
	}

	if err = db.Seed(ctx); err != nil {
		return fmt.Errorf("db seed failed: %w", err)
	}

	categories, err := db.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("list categories failed: %w", err)
	}
	for _, category := range categories {
		fmt.Println(category)
	}

	movies, err := db.ListMovies(ctx)
	if err != nil {
		return fmt.Errorf("list movies failed: %w", err)
	}

	for _, movie := range movies {
		fmt.Println(movie.String())
	}

	return nil
}

func pprint(v interface{}) {
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(s))
}

func main() {
	fmt.Println("ok")

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := App{log: log}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}



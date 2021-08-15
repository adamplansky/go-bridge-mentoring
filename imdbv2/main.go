package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/adamplansky/go-bridge-mentoring/imdbv2/models"

	"github.com/adamplansky/go-bridge-mentoring/imdbv2/persistance"
	"go.uber.org/zap"
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

	resp, err := db.CreateCategory(ctx, models.Category{Title: "title1"})
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}
	a.log.Info("create category", zap.String("response", resp.String()))

	categories, err := db.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("list categories failed: %w", err)
	}
	a.log.Info("list categories")
	for _, category := range categories {
		fmt.Printf("category.Title: %s, category.ID: %s\n",
			category.Title,
			category.ID,
		)

		cat, err := db.GetCategory(ctx, category.ID)
		if err != nil {
			return fmt.Errorf("get category failed: %w", err)
		}
		fmt.Printf("[GET CATEGORY]: category.Title: %s, category.ID: %s\n",
			cat.Title,
			cat.ID,
		)
	}

	//
	//link := Link{
	//	URL:   "https://github.com/dgraph-io/dgo",
	//	DType: []string{"Link"},
	//}
	//
	//lb, err := json.Marshal(link)
	//if err != nil {
	//	return fmt.Errorf("failed to marshal %w", err)
	//}
	//
	//mu := &api.Mutation{
	//	SetJson: lb,
	//}
	//res, err := txn.Mutate(ctx, mu)
	//if err != nil {
	//	return fmt.Errorf("failed to mutate %w", err)
	//}
	//pprint(res)

	q := `
{
	q(func: has(User.name)){
		User.uid
		User.name
		User.age
  }
}`

	res, err := db.NewReadOnlyTxn().Query(ctx, q)
	if err != nil {
		return fmt.Errorf("query failed %w", err)
	}

	var r struct {
		People []struct {
			Uid  string `json:"User.uid,omitempty"`
			Name string `json:"User.name,omitempty"`
			Age  int    `json:"User.age,omitempty"`
		} `json:"q"`
	}

	err = json.Unmarshal(res.Json, &r)
	if err != nil {
		return err
	}

	fmt.Println(r)
	for _, p := range r.People {
		fmt.Printf("\n--------\n%+v\n", p)
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

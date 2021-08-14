package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adamplansky/go-bridge-mentoring/imdbv2/persistance"
	"go.uber.org/zap"
	"log"
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

func (a App) Run(ctx context.Context) error {
	db, err := persistance.NewDB("localhost:9080")
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer db.Close()

	//txn := db.NewTxn()
	//defer txn.Commit(ctx)
	//

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
	query1(func: has(name)){
		uid
		age
		name
  	}
}`

	res, err := db.NewReadOnlyTxn().Query(ctx, q)
	if err != nil {
		return fmt.Errorf("query failed %w", err)
	}
	fmt.Printf("%s\n", res.Json)

	type Person struct {
		Uid      string     `json:"uid,omitempty"`
		Name     string     `json:"name,omitempty"`
		Age      int        `json:"age,omitempty"`
	}

	var r struct{
		People []Person `json:"query1"`
	}

	err = json.Unmarshal(res.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range r.People {
		fmt.Printf("\n--------\n%+v\n", p)
	}



	return nil
}

func pprint(v interface{}){
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(s))
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := App{log: log}

	if err := app.Run(ctx); err != nil {
		log.Fatal()
	}

}

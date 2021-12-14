package persistance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v210/protos/api"

	"github.com/adamplansky/go-bridge-mentoring/imdbv2/models"
)


func (d *DB) CreateCategory(
	ctx context.Context,
	category models.Category,
) (*api.Response, error) {
	b, err := json.Marshal(category)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return d.MutateTx(ctx, b)
}

func (d *DB) ListCategories(ctx context.Context) ([]models.Category, error) {
	const q = `
	{
  		q(func: has(Category.title)){
			uid
			Category.title
	  	}
	}`
	resp, err := d.NewReadOnlyTxn().Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	var r struct {
		Categories []models.Category `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return r.Categories, nil
}

func (d *DB) GetCategory(ctx context.Context, uid string) (*models.Category, error) {
	variables := map[string]string{"$id": uid}
	const q = `
	query getCategory($id: string) {
	    q(func: uid($id)) {
		uid
		Category.title
	    }
	}`

	resp, err := d.NewReadOnlyTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	fmt.Printf("resp json: %s\n", resp.Json)

	var r struct {
		Category []models.Category `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return &r.Category[0], nil
}

// filter movies by the person

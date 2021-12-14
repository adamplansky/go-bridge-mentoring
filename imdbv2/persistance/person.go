package persistance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adamplansky/go-bridge-mentoring/imdbv2/models"
	"github.com/dgraph-io/dgo/v210/protos/api"
)


func (d *DB) CreatePerson(
	ctx context.Context,
	person models.Person,
) (*api.Response, error) {
	b, err := json.Marshal(person)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return d.MutateTx(ctx, b)
}

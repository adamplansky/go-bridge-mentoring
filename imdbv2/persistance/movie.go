package persistance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adamplansky/go-bridge-mentoring/imdbv2/models"
	"github.com/dgraph-io/dgo/v210/protos/api"
)


func (d *DB) CreateMovie(
	ctx context.Context,
	movie models.Movie,
) (*api.Response, error) {
	b, err := json.Marshal(movie)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return d.MutateTx(ctx, b)
}

func (d *DB) ListMovies(ctx context.Context) ([]models.Movie, error) {
	const q = `
	{
  		q(func: has(Movie.title)){
			Movie.title
			Movie.release_date
			Movie.score
			Movie.description
			Movie.duration
			Movie.categories {
				Category.title
			}
			Movie.artists {
				Person.full_name
				Person.born
				Person.description
			}
			Movie.directors {
				Person.full_name
				Person.born
				Person.description
			}
		}
	}`
	resp, err := d.NewReadOnlyTxn().Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	var r struct {
		Movies []models.Movie `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return r.Movies, nil
}
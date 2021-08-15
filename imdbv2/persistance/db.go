package persistance

import (
	"context"
	"fmt"

	"google.golang.org/grpc/encoding/gzip"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"google.golang.org/grpc"
)

type DB struct {
	Conn *grpc.ClientConn
	*dgo.Dgraph
}

// NewDB creates new DGraph database connection. If key is provided it is dgraph cloud
// connection, otherwise it is insecure grpc connection to localhost.
func NewDB(dsn, key string) (*DB, error) {
	var err error
	var conn *grpc.ClientConn

	if key != "" {
		conn, err = dgo.DialSlashEndpoint(dsn, key)
	} else {
		conn, err = grpc.Dial(dsn,
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("grpc dial failed: %w", err)
	}

	return &DB{
		Conn:   conn,
		Dgraph: dgo.NewDgraphClient(api.NewDgraphClient(conn)),
	}, nil

}

func (d *DB) Migrate(ctx context.Context, cleanDB bool) error {
	op := &api.Operation{}
	if cleanDB {
		op.DropOp = api.Operation_ALL
	}
	op.Schema = `
		title: string @index(exact) .
		release_date: datetime .
		score: int .
		description: string .
		full_name: string .

		type Movie {
			title: string
			release_date: datetime
			score: int
			description: string
		}
		type Artist {
			full_name: string
		}
		type Category {
			title: string
		}
	`

	if err := d.Dgraph.Alter(ctx, op); err != nil {
		return fmt.Errorf("dgraph alter scheme failed: %w", err)
	}
	return nil
}

func (d *DB) Seed(ctx context.Context) error {
	err := d.Dgraph.Alter(ctx, &api.Operation{
		Schema: `
			name: string @index(term) .
			balance: int .
		`,
	})
	if err != nil {
		return fmt.Errorf("dgraph alter scheme failed: %w", err)
	}
	return nil
}

func (d *DB) MutateTx(ctx context.Context, b []byte) (*api.Response, error) {
	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   b,
	}

	response, err := d.Dgraph.NewTxn().Mutate(ctx, mu)
	if err != nil {
		return nil, fmt.Errorf("txn mutate failed: %w", err)
	}
	return response, nil
}

func (d *DB) Close() error {
	return d.Conn.Close()
}

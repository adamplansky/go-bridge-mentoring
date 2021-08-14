package persistance

import (
	"fmt"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"google.golang.org/grpc"
)

type DB struct {
	Conn *grpc.ClientConn
	*dgo.Dgraph
}
func NewDB(dsn string) (*DB, error){
	conn, err := grpc.Dial(dsn, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc dial failed: %w", err)
	}

	return &DB{
		Conn:   conn,
		Dgraph: dgo.NewDgraphClient(api.NewDgraphClient(conn)),
	}, nil

}

func (d *DB) Close() {
	d.Conn.Close()
}

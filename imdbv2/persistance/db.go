package persistance

import (
	"context"
	"fmt"
	"github.com/adamplansky/go-bridge-mentoring/imdbv2/models"
	"time"

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
		Movie.title: string @index(exact) .
		Movie.release_date: datetime .
		Movie.score: int .
		Movie.description: string .
		Movie.duration: int .

		Category.title: string @index(exact) .

		Person.full_name: string .
		Person.born: datetime .
		Person.description: string .

		type Movie {
			Movie.title: string
			Movie.release_date: datetime
			Movie.score: int
			Movie.description: string
			Movie.duration: int

			Movie.categories: [Category]
			Movie.artists: [Person]
			Movie.directors: [Person]
		}
		type Person {
			Person.full_name: string
			Person.born: datetime
			Person.description: string

			Person.artists_movies: [Movie]
			Person.directors_movies: [Movie]
		}
		type Category {
			Category.title: string
			Category.movies: [Movie]
		}
	`

	if err := d.Dgraph.Alter(ctx, op); err != nil {
		return fmt.Errorf("dgraph alter scheme failed: %w", err)
	}
	return nil
}

func (d *DB) Seed(ctx context.Context) error {
	catDrama := models.Category{Title: "drama"}
	_, err := d.CreateCategory(ctx, catDrama)
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}

	catRomance := models.Category{Title: "romance"}
	_, err = d.CreateCategory(ctx, catRomance)
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}

	catCrime := models.Category{Title: "crime"}
	_, err = d.CreateCategory(ctx, catCrime)
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}

	catFantasy := models.Category{Title: "fantasy"}
	_, err = d.CreateCategory(ctx, catFantasy)
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}

	catMystery := models.Category{Title: "catMystery"}
	_, err = d.CreateCategory(ctx, catMystery)
	if err != nil {
		return fmt.Errorf("create category failed: %w", err)
	}

	personHanks := models.Person{
		FullName:    "Tom Hanks",
		Born:        timeMustParse("1956-07-09"),
		Description: "Thomas Jeffrey Hanks was born in Concord, California, to Janet Marylyn (Frager), a hospital worker, and Amos Mefford Hanks, an itinerant cook. His mother's family, originally surnamed \"Fraga\", was entirely Portuguese, while his father was of mostly English ancestry. Tom grew up in what he has called a \"fractured\" family. He moved around a great ...",
	}

	personDarabont := models.Person{
		FullName:    "Frank Darabont",
		Born:        timeMustParse("1959-01-28"),
		Description: "Frank Darabont...",
	}

	personMorse := models.Person{
		FullName:    "David Morse",
		Born:        timeMustParse("1953-10-11"),
		Description: "David Morse was born on November 12, 1975 in Pittsburgh, Pennsylvania, USA. He is known for his work on The Screening (2007), Sofia for Now (2006) and The Chair (2021). ",
	}

	personPitt := models.Person{
		FullName:    "Brad Pitt",
		Born:        timeMustParse("1963-12-18"),
		Description: "An actor and producer known as much for his versatility as he is for his handsome face, Golden Globe-winner Brad Pitt's most widely recognized role may be Tyler Durden in Fight Club (1999). However, his portrayals of Billy Beane in Moneyball (2011), and Rusty Ryan in the remake of Ocean's Eleven (2001) and its sequels, also loom large in his ..",
	}

	personFreeman := models.Person{
		FullName:    "Morgan Freeman",
		Born:        timeMustParse("1937-06-01"),
		Description: "With an authoritative voice and calm demeanor, this ever popular American actor has grown into one of the most respected figures in modern US cinema. Morgan was born on June 1, 1937 in Memphis, Tennessee, to Mayme Edna (Revere), a teacher, and Morgan Porterfield Freeman, a barber. The young Freeman attended Los Angeles City College before serving ..",
	}

	personFincher := models.Person{
		FullName:    "David Fincher",
		Born:        timeMustParse("1962-08-28"),
		Description: "With an authoritative voice and calm demeanor, this ever popular American actor has grown into one of the most respected figures in modern US cinema. Morgan was born on June 1, 1937 in Memphis, Tennessee, to Mayme Edna (Revere), a teacher, and Morgan Porterfield Freeman, a barber. The young Freeman attended Los Angeles City College before serving ..",
	}

	personZemeckis := models.Person{
		FullName:    "Robert Zemeckis",
		Born:        timeMustParse("1951-05-14"),
		Description: "A whiz-kid with special effects, Robert is from the Spielberg camp of film-making (Steven Spielberg produced many of his films). Usually working with writing partner Bob Gale, Robert's earlier films show he has a talent for zany comedy (Romancing the Stone (1984), 1941 (1979)) and special effect vehicles (Who Framed Roger Rabbit (1988) and Back to...",
	}

	// ----------------------------------------- movies
	movForrest := models.Movie{
		Title:       "Forrest Gump",
		ReleaseDate: timeMustParse("1994-07-06"),
		Categories:  []models.Category{catDrama, catRomance},
		Duration:    142,
		Score:       94,
		Description: "The presidencies of Kennedy and Johnson, the Vietnam War, the Watergate scandal and other historical events unfold from the perspective of an Alabama man with an IQ of 75, whose only desire is to be reunited with his childhood sweetheart.",
		Directors:   []models.Person{personZemeckis},
		Artists:     []models.Person{personHanks},
		Keywords:    nil,
		Comments:    nil,
	}
	resp, err := d.CreateMovie(ctx, movForrest)
	if err != nil {
		return fmt.Errorf("create movie failed: %w", err)
	}

	movGreenMile := models.Movie{
		Title:       "The Green Mile",
		ReleaseDate: timeMustParse("1999-12-06"),
		Categories:  []models.Category{catDrama, catCrime, catFantasy},
		Duration:    188,
		Score:       93,
		Description: "The lives of guards on Death Row are affected by one of their charges: a black man accused of child murder and rape, yet who has a mysterious gift.",
		Directors:   []models.Person{personDarabont},
		Artists:     []models.Person{personHanks, personMorse},
		Keywords:    nil,
		Comments:    nil,
	}
	resp, err = d.CreateMovie(ctx, movGreenMile)
	if err != nil {
		return fmt.Errorf("create movie failed: %w", err)
	}

	movSeven := models.Movie{
		Title:       "Se7en",
		ReleaseDate: timeMustParse("1995-09-22"),
		Categories:  []models.Category{catDrama, catCrime, catMystery},
		Duration:    127,
		Score:       92,
		Description: "Two detectives, a rookie and a veteran, hunt a serial killer who uses the seven deadly sins as his motives.",
		Directors:   []models.Person{personFincher},
		Artists:     []models.Person{personPitt, personFreeman},
		Keywords:    nil,
		Comments:    nil,
	}
	resp, err = d.CreateMovie(ctx, movSeven)
	if err != nil {
		return fmt.Errorf("create movie failed: %w", err)
	}

	_ = resp

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

const shortForm = "2006-02-02"

func timeMustParse(dateStr string) time.Time {
	t, err := time.Parse(shortForm, dateStr)
	if err != nil {
		panic(err)
	}
	return t
}
package roundtripper

import (
	"log"
	"net/http"
	"os"
)

type myRoundTripper struct {
	l *log.Logger
}

func New() http.RoundTripper {
	return &myRoundTripper{l: log.New(os.Stdout, "New: ", log.Ldate)}
}


func (mrt *myRoundTripper) RoundTrip(r *http.Request) (*http.Response, error){
	mrt.l.Println("req: http method: ", r.Method)
	mrt.l.Printf("req: headers %+v", r.Header)


	t := http.Transport{}
	resp, err :=  t.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	mrt.l.Println("resp: http code: ", resp.Status)
	mrt.l.Printf("resp: headers %+v", r.Header)


	return resp, nil
}
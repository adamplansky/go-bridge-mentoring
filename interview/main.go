package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/semaphore"
)

type service struct {
	queue chan int
	wg sync.WaitGroup
}

func NewService() *service {
	return &service{
		queue:         make(chan int, 5),
	}
}

func (s *service) callbackH() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", 400)
			return
		}
		var IDS []int
		err := json.NewDecoder(r.Body).Decode(&IDS)
		if err != nil {
			http.Error(w, "cannot decode ids", 500)
			return
		}
		for _, ID := range IDS {
			s.queue <- ID
		}
	}
}

func (s *service) Close() {
	close(s.queue)
	s.wg.Wait()
}



func (s *service) worker() {
	s.wg.Add(1)
	defer s.wg.Done()

	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(20))

	for ID := range s.queue {
		// sem.Acquire cannot fail because of context.Backround()
		_ = sem.Acquire(context.Background(), 1)

		URL := fmt.Sprintf("http://localhost:8081/%d", ID)
		eg.Go(func() error{
			defer sem.Release(1)
			c := http.Client{
				Timeout: 4 * time.Second,
			}

			req, err := http.NewRequest(http.MethodGet, URL, nil)
			if err != nil {
				log.Println(err)
				return nil
			}

			resp, err := c.Do(req)
			if err != nil {
				log.Println(err)
				return nil
			}

			code := resp.StatusCode
			if code == http.StatusOK {
				s.saveToDb(ID)
			}
			return nil
		})
	}
	// wait for all go routines to stop
	_ = eg.Wait()

}

func (s *service) saveToDb(ID int) {
	// save into db
	fmt.Println(ID, time.Now())
}

// dbCleanup every 30 second cleanup all record older than 2 days
func (s *service) dbCleanup(ctx context.Context) {
	s.wg.Add(1)
	defer s.wg.Done()

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		fmt.Println("deleting items older then 2 days")
		time.Sleep(1 * time.Hour)
	}

}

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelFn()

	s := NewService()
	go s.worker()
	go s.dbCleanup(ctx)
	defer s.Close()

	m := http.NewServeMux()
	m.HandleFunc("/", s.callbackH())

	server := http.Server{
		Addr:              ":8080",
		Handler:           m,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal(err)
	}

}

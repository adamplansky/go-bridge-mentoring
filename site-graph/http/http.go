package http

import (
	"context"
	"log"
	"net"
	"net/http"
)

func ServeHTTP(ctx context.Context, addr string, h http.Handler) {
	httpS := http.Server{
		Addr:    addr,
		Handler: h,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		if err := httpS.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
}

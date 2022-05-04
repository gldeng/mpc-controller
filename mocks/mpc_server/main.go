package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	err := startServer()
	if err != nil {
		log.Fatal(err)
	}
}

func startServer() error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("."))
	})

	webrpcHandler := NewMPCServiceServer(&MPCServiceRPC{})
	r.Handle("/*", webrpcHandler)

	return http.ListenAndServe(":4242", r)
}

type mpcService interface {
	keygen(ctx context.Context) error
	sign(ctx context.Context) error
	result(ctx context.Context) error
}

type MPCServiceRPC struct {
}

func (s *MPCServiceRPC) Keygen(ctx context.Context, keygenReq *KeygenRequest) error {
	return nil
}

func (s *MPCServiceRPC) Sign(ctx context.Context, signReq *SignRequest) error {
	return nil
}

func (s *MPCServiceRPC) Result(ctx context.Context, reqID string) (*Result, error) {
	return nil, nil
}

package prom

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

type MetricsService struct {
	Ctx       context.Context
	ServeAddr string
}

// todo: add Close() eror method

func (p *MetricsService) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Addr:    p.ServeAddr,
		Handler: http.DefaultServeMux,
		//TLSConfig: nil, todo: more fields to consider
	}

	wg := &sync.WaitGroup{}
	var err error
	go func() {
		wg.Add(1)
		err = srv.ListenAndServe()
		wg.Done()
	}()
	if err != nil {
		return errors.Wrapf(err, "got an error to start MetricsService")
	}

	<-p.Ctx.Done()
	if err = srv.Shutdown(context.Background()); err != nil {
		err = errors.Wrapf(err, "got an error to shutdown MetricsService")
	}
	wg.Wait()
	return err
}

func (p *MetricsService) Close() error {
	// todo:
	return nil
}

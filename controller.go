package mpc_controller

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/services/group"
	"github.com/avalido/mpc-controller/services/keygen"
	"github.com/avalido/mpc-controller/services/reward"
	"github.com/avalido/mpc-controller/services/stake"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

type MpcController struct {
	ID       string
	Services []MpcControllerService
}

func (c *MpcController) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, service := range c.Services {
		g.Go(func() error {
			return service.Start(ctx)
		})
	}

	fmt.Printf("%v services started.\n", c.ID)
	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

	fmt.Printf("%v services closed.\n", c.ID)
	return nil
}

// todo: add concrete services and necessary configs

func RunMpcController(c *cli.Context) error {
	// Handle graceful shutdown.
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		shutdown()
	}()

	// Run the mpc-controller
	controller := MpcController{
		Services: []MpcControllerService{
			&group.Group{},
			&keygen.Keygen{},
			&stake.Manager{},
			&reward.Reward{},
		},
	}

	if err := controller.Run(shutdownCtx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/config"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/urfave/cli/v3"
	"log/slog"
	"runtime"
)

func ListenCommand() *cli.Command {
	return &cli.Command{
		UseShortOptionHandling: true,
		Name:                   "listen",
		Usage:                  "Listen to CDC events without sending them to any output",
		Flags:                  []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, ok := ctx.Value(ContextValueConfig).(*config.Config)
			if !ok {
				return fmt.Errorf("config not found in context")
			}
			return Listen(ctx, cfg)
		},
	}
}

// Listen starts the entire CDC listening process, including data source, storage, and output handling
func Listen(ctx context.Context, config *config.Config) error {
	iSource, iStore, iOutput, err := SetupComponents(config)
	if err != nil {
		return err
	}
	defer CloseSetupComponents(iSource, iStore, iOutput)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sourceErrChan := startSource(ctx, iSource)
	workerCount := runtime.NumCPU()
	startWorkers(ctx, iSource, iOutput, workerCount)
	if err := waitSourceError(sourceErrChan); err != nil {
		slog.Error("source error detected, shutting down", "error", err)
		return err
	}
	slog.Info("CDC process completed successfully")
	return nil
}

// Start Source and return the error channel
func startSource(ctx context.Context, iSource source.ISource) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		slog.Info("starting source goroutine")
		if err := iSource.Run(ctx); err != nil {
			slog.Error("source run failed", "error", err)
			errChan <- err
		}
		close(errChan)
	}()
	return errChan
}

// Start the worker pool
func startWorkers(ctx context.Context, iSource source.ISource, iOutput output.IOutput, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go workerLoop(ctx, i, iSource, iOutput)
	}
	slog.Info("started all workers", "count", workerCount)
}

// Worker main loop
func workerLoop(ctx context.Context, id int, iSource source.ISource, iOutput output.IOutput) {
	slog.Info("worker started", "workerID", id)
	for {
		select {
		case event, ok := <-iSource.GetChanEventData():
			if !ok {
				slog.Info("event channel closed", "workerID", id)
				return
			}
			if err := output.SendWithRetry(ctx, iOutput, event, 3); err != nil {
				slog.Error("failed to send event", "workerID", id, "error", err)
			}
		case <-ctx.Done():
			slog.Info("context canceled, worker exiting", "workerID", id)
			return
		}
	}
}

// Waiting for data source error signal
func waitSourceError(errChan <-chan error) error {
	if err, ok := <-errChan; ok && err != nil {
		return err
	}
	return nil
}

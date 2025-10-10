package cmd

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/config"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
	"runtime"
	"sync"
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
	// Create Store instance
	iStore, err := store.NewStore(config.Store)
	if err != nil {
		return err // Return error if Store creation fails
	}
	// Create Source instance
	iSource, err := source.NewSource(config.Source)
	if err != nil {
		return err // Return error if Source creation fails
	}
	// Inject Store into Source
	iSource.WithStore(iStore)
	// Create Output instance
	iOutput, err := output.NewOutput(config.Output)
	if err != nil {
		return err // Return error if Output creation fails
	}

	// Create cancellable context for controlling goroutine lifecycle
	ctx, cancel := context.WithCancel(ctx)
	// Define a callback function to close resources uniformly
	closeCallback := func() {
		slog.Info("closing source and output")
		if err := iSource.Close(); err != nil {
			slog.Error("failed to close source", "error", err)
		}
		if err := iOutput.Close(); err != nil {
			slog.Error("failed to close output", "error", err)
		}
		cancel() // Cancel the context
	}

	defer closeCallback() // Automatically call close callback when function exits

	// Create a channel to monitor errors in the data source
	sourceErrChan := make(chan error, 1)

	// Start a goroutine to run the source
	go func() {
		slog.Info("starting source run goroutine")
		if err := iSource.Run(ctx); err != nil {
			// Log error and send it to the channel if source run fails
			slog.Error("source run failed", "error", err)
			sourceErrChan <- err
		}
		close(sourceErrChan) // Close the channel after the source finishes
	}()

	// Define the number of workers, using the number of CPU cores
	workerCount := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Start worker goroutines to handle data source events
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			defer wg.Done() // Notify WaitGroup when the goroutine finishes
			slog.Info("worker started", "workerID", id)
			for {
				select {
				case event, ok := <-iSource.GetChanEventData():
					// Read events from the channel
					if !ok {
						// If the channel is closed, worker exits
						slog.Info("event channel closed", "workerID", id)
						return
					}
					slog.Info("CDC Event", slog.Any("event", event))
					// Send event to output
					if err := output.SendWithRetry(ctx, iOutput, event, 3); err != nil {
						slog.Error("failed to send event", "workerID", id, "error", err)
					}
				case <-ctx.Done():
					// If context is canceled, worker exits
					slog.Info("context canceled, worker exiting", "workerID", id)
					return
				}
			}
		}(i)
	}
	// Wait for data source errors or worker completion
	err, ok := <-sourceErrChan
	if ok && err != nil {
		// If source run failed, close all resources and wait for workers to exit
		closeCallback()
		wg.Wait()
		slog.Error("exiting program due to source run error", "error", err)
		os.Exit(1)
	}
	// If source run completed successfully, wait for workers to finish
	wg.Wait()
	return nil
}

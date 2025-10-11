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
	"runtime"
	"sync"
)

func ListenCommand() *cli.Command {
	return &cli.Command{
		UseShortOptionHandling: true,
		Name:                   "listen",
		Usage:                  "Listen to CDC events without sending them to any output",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "processor",
				Aliases: []string{"p"},
				Value:   runtime.NumCPU(),
				Usage:   "Number of processors to handle CDC events",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, ok := ctx.Value(ContextValueConfig).(*config.Config)
			if !ok {
				return fmt.Errorf("config not found in context")
			}
			processorCount := cmd.Int("processor")
			if processorCount <= 0 {
				processorCount = runtime.NumCPU()
			}
			return Listen(ctx, processorCount, cfg)
		},
	}
}

// Listen starts the entire CDC listening process, including data source, storage, and output handling
func Listen(ctx context.Context, processorCount int, cfg *config.Config) error {
	// Initialize store
	iStore, err := store.NewStore(cfg.Store)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Initialize source
	iSource, err := source.NewSource(cfg.Source)
	if err != nil {
		_ = iStore.Close()
		return fmt.Errorf("failed to initialize source: %w", err)
	}
	iSource.WithStore(iStore)

	// Initialize output
	iOutput, err := output.NewOutput(cfg.Output)
	if err != nil {
		_ = iSource.Close()
		_ = iStore.Close()
		return fmt.Errorf("failed to initialize output: %w", err)
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)

	// Ensure resources are closed only once
	var once sync.Once
	closeCallback := func() {
		once.Do(func() {
			slog.Info("closing source and output")
			_ = iSource.Close()
			_ = iOutput.Close()
			cancel()
		})
	}
	defer closeCallback()

	// Channel to capture source errors
	sourceErrChan := make(chan error, 1)

	// Run source in a goroutine
	go func() {
		slog.Info("starting source run goroutine")
		if err := iSource.Run(ctx); err != nil {
			slog.Error("source run failed", "error", err)
			select {
			case sourceErrChan <- err:
			default:
				slog.Warn("source error channel full, dropping error")
			}
		}
		close(sourceErrChan)
	}()

	// Start processor goroutines
	var wg sync.WaitGroup
	wg.Add(processorCount)

	for i := 0; i < processorCount; i++ {
		go func(id int) {
			defer wg.Done()
			slog.Info("processor started", "id", id)
			for event := range iSource.GetChanEventData() {
				slog.Debug("CDC Event", slog.Any("event", event))
				if err := output.SendWithRetry(ctx, iOutput, event, 3); err != nil {
					slog.Error("failed to send event", "processorID", id, "error", err)
				}
			}
			slog.Info("processor exiting", "id", id)
		}(i)
	}

	// Wait for source errors or context cancellation
	select {
	case err := <-sourceErrChan:
		if err != nil {
			closeCallback()
			wg.Wait()
			return fmt.Errorf("source run failed: %w", err)
		}
	case <-ctx.Done():
		slog.Info("context canceled, shutting down")
	}

	// Wait for all processors to finish
	wg.Wait()
	return nil
}

package cmd

import (
	"fmt"
	"github.com/chihqiang/dbxgo/config"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"log/slog"
)

// SetupComponents components: Store, Source, Output
func SetupComponents(cfg *config.Config) (source.ISource, store.IStore, output.IOutput, error) {
	iStore, err := store.NewStore(cfg.Store)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create store: %w", err)
	}
	iSource, err := source.NewSource(cfg.Source)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create source: %w", err)
	}
	iSource.WithStore(iStore)
	iOutput, err := output.NewOutput(cfg.Output)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create output: %w", err)
	}
	return iSource, iStore, iOutput, nil
}

// CloseSetupComponents resources uniformly
func CloseSetupComponents(iSource source.ISource, iStore store.IStore, iOutput output.IOutput) {
	slog.Info("closing source and output")
	if err := iSource.Close(); err != nil {
		slog.Error("failed to close source", "error", err)
	}
	if err := iStore.Close(); err != nil {
		slog.Error("failed to close store", "error", err)
	}
	if err := iOutput.Close(); err != nil {
		slog.Error("failed to close output", "error", err)
	}
}

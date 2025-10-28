package cmd

import (
	"fmt"
	"chihqiang/dbxgo/config"
	"chihqiang/dbxgo/output"
	"chihqiang/dbxgo/pkg/logx"
	"chihqiang/dbxgo/source"
	"chihqiang/dbxgo/store"
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
	logx.Info("closing source and output")
	if err := iSource.Close(); err != nil {
		logx.Error("failed to close source: %v", err)
	}
	if err := iStore.Close(); err != nil {
		logx.Error("failed to close store: %v", err)
	}
	if err := iOutput.Close(); err != nil {
		logx.Error("failed to close output: %v", err)
	}
}

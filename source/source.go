package source

import (
	"context"
	"fmt"
	"chihqiang/dbxgo/store"
	"chihqiang/dbxgo/types"
)

const (
	StoreKeyPosition = "_dbxgo_position"
)

type SourceType string

var (
	SourceTypeMysql SourceType = "mysql"
)

// Config Defines the data source configuration structure
// Used to configure database connection information and storage settings
type Config struct {
	// Type The type of the data source, such as "mysql"
	Type  SourceType   `yaml:"type" json:"type" mapstructure:"type" env:"SOURCE_TYPE,required"`
	Mysql MysqlConfig  `yaml:"mysql" json:"mysql" mapstructure:"mysql"`
	Store store.Config `yaml:"store" json:"store" mapstructure:"store"`
}

// ISource Defines the data source interface
// All specific data source implementations must implement this interface
type ISource interface {
	// WithStore Sets the storage interface, used for persisting or reading offsets and states
	WithStore(store store.IStore)

	// Run Starts the data source listener
	// ctx: The context, used to control cancellation and timeout
	// Return value: Possible error
	Run(ctx context.Context) error

	// GetChanEventData Returns the event data channel
	// External systems use this channel to receive database change events
	// Return value: A read-only event data channel
	GetChanEventData() <-chan types.EventData

	// Close Closes the data source and releases resources
	// Return value: Possible error
	Close() error
}

// NewSource Creates the corresponding data source instance based on the configuration
// cfg: The data source configuration information
// Return value: The data source interface implementation and possible errors
func NewSource(cfg Config) (ISource, error) {
	switch cfg.Type {
	case SourceTypeMysql:
		return NewMySQLSource(cfg.Mysql)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}

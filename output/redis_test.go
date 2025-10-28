package output

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"chihqiang/dbxgo/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisOutput(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cfg := RedisConfig{
		Addr: mr.Addr(),
		DB:   0,
		Key:  "test-key",
	}

	rout, err := NewRedisOutput(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, rout)
	assert.Equal(t, cfg.Key, rout.key)
	err = rout.Close()
	assert.NoError(t, err)
}

func TestRedisOutput_Send(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cfg := RedisConfig{
		Addr: mr.Addr(),
		DB:   0,
		Key:  "test-events",
	}

	rout, err := NewRedisOutput(cfg)
	assert.NoError(t, err)
	defer rout.Close()

	ctx := context.Background()

	event := types.EventData{
		Time:     time.Now(),
		ServerID: 12345,
		Pos:      6789,
		Row: types.EventRowData{
			Time:     time.Now().UnixMilli(),
			Database: "testdb",
			Table:    "users",
			Type:     types.InsertEventRowType,
			Data: map[string]any{
				"id":   1,
				"name": "Alice",
			},
		},
	}

	err = rout.Send(ctx, event)
	assert.NoError(t, err)

	values, _ := mr.List(cfg.Key)
	assert.Len(t, values, 1, "Redis list should contain one item")

	var stored types.EventData
	err = json.Unmarshal([]byte(values[0]), &stored)
	assert.NoError(t, err)
	assert.Equal(t, event.ServerID, stored.ServerID)
	assert.Equal(t, event.Row.Table, stored.Row.Table)
	assert.Equal(t, event.Row.Data["name"], stored.Row.Data["name"])
}

func TestRedisOutput_Send_InvalidJSON(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cfg := RedisConfig{
		Addr: mr.Addr(),
		DB:   0,
		Key:  "bad-events",
	}

	rout, err := NewRedisOutput(cfg)
	assert.NoError(t, err)
	defer rout.Close()

	ctx := context.Background()

	event := types.EventData{
		Time:     time.Now(),
		ServerID: 999,
		Pos:      111,
		Row: types.EventRowData{
			Time:     time.Now().UnixMilli(),
			Database: "bad_db",
			Table:    "weird_table",
			Type:     types.UpdateEventRowType,
			Data: map[string]any{
				"broken": make(chan int),
			},
		},
	}

	err = rout.Send(ctx, event)
	assert.Error(t, err, "should return error for invalid JSON")
}

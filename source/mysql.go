package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/store"
	"github.com/chihqiang/dbxgo/types"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-mysql-org/go-mysql/schema"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

var (
	// DefaultMysqlExcludeTableRegex 默认排除的系统库表正则
	// 这些库表一般为 MySQL 内置系统库，无需采集或处理
	DefaultMysqlExcludeTableRegex = []string{
		"mysql.*",              // MySQL 系统库
		"information_schema.*", // 信息架构库
		"performance_schema.*", // 性能监控库
		"sys.*",                // 系统视图库
	}
)

// MysqlConfig MySQL 配置实体
// 用于描述 MySQL 数据源的连接信息和表过滤规则。
// 表筛选逻辑：
//  1. 优先根据 ExcludeTableRegex 进行排除
//  2. 若 IncludeTableRegex 不为空，则仅包含匹配的表
//  3. 如果两者都为空，则默认处理全部表
type MysqlConfig struct {
	Addr              string   `yaml:"addr" json:"addr" mapstructure:"addr" env:"SOURCE_MYSQL_ADDR" envDefault:"127.0.0.1:3306"`
	User              string   `yaml:"user" json:"user" mapstructure:"user" env:"SOURCE_MYSQL_USER" envDefault:"root"`
	Password          string   `yaml:"password" json:"password" mapstructure:"password" env:"SOURCE_MYSQL_PASSWORD" envDefault:""`
	ExcludeTableRegex []string `yaml:"exclude_table_regex" json:"exclude_table_regex" mapstructure:"exclude_table_regex" env:"SOURCE_MYSQL_EXCLUDE_TABLE_REGEX"`
	IncludeTableRegex []string `yaml:"include_table_regex" json:"include_table_regex" mapstructure:"include_table_regex" env:"SOURCE_MYSQL_INCLUDE_TABLE_REGEX"`
}

// MySQLSource 是MySQL数据源的具体实现
// 负责连接MySQL数据库，监听binlog事件并转换为统一的事件格式
type MySQLSource struct {
	// mu 用于保证并发安全的互斥锁
	mu sync.Mutex
	// 嵌入 canal.DummyEventHandler 以实现事件处理接口
	canal.DummyEventHandler
	// store 存储接口，用于持久化或读取偏移量及状态
	store store.IStore
	// canal MySQL binlog 解析器实例
	canal *canal.Canal
	// cfg 数据源配置信息
	cfg MysqlConfig
	// eventDataChan 事件数据输出通道
	eventDataChan chan types.EventData
	// running 表示当前数据源是否正在运行
	running bool
}

// MysqlPosition 定义MySQL binlog位置信息结构
// 用于保存和恢复同步位置
type MysqlPosition struct {
	// File binlog文件名
	File string `json:"file"`
	// Pos binlog文件中的位置偏移量
	Pos uint32 `json:"pos"`
}

// NewMySQLSource 创建一个MySQL数据源实例
// cfg: MySQL数据源配置信息
// 返回值: 实现了ISource接口的MySQLSource实例和可能的错误
func NewMySQLSource(cfg MysqlConfig) (ISource, error) {
	if len(cfg.ExcludeTableRegex) == 0 {
		cfg.ExcludeTableRegex = DefaultMysqlExcludeTableRegex
	}
	cc := canal.NewDefaultConfig()
	// 设置数据库连接信息
	cc.Addr = cfg.Addr
	cc.User = cfg.User
	cc.Password = cfg.Password
	// 不需要mysqldump执行路径，使用纯Go实现
	cc.Dump.ExecutionPath = ""
	cc.ExcludeTableRegex = cfg.ExcludeTableRegex
	if len(cfg.IncludeTableRegex) > 0 {
		cc.IncludeTableRegex = cfg.IncludeTableRegex
	}
	// 创建MySQLSource实例
	source := &MySQLSource{
		cfg:           cfg,
		eventDataChan: make(chan types.EventData, 10240), // 缓冲区大小为10240的事件通道
	}
	// 创建canal实例
	c, err := canal.NewCanal(cc)
	if err != nil {
		return nil, err
	}

	// 保存canal实例并设置事件处理器
	source.canal = c
	source.canal.SetEventHandler(source)
	return source, nil
}
func (s *MySQLSource) WithStore(store store.IStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = store
}

// Run 启动MySQL数据源监听
// ctx: 上下文，用于控制取消和超时
// 返回值: 可能的错误
func (s *MySQLSource) Run(ctx context.Context) error {
	if s.store == nil {
		return fmt.Errorf("store is not initialized, cannot run MySQLSource")
	}
	// 检查是否已经在运行
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("it is already running")
	}
	s.running = true
	s.mu.Unlock()
	// 加载上次保存的同步位置
	startPos := s.loadPosition()
	// 创建一个用于接收canal退出通知的通道
	done := make(chan error, 1)

	// 启动canal监听（在后台goroutine中）
	go func() {
		// 使用RunFrom方法从指定位置开始同步binlog
		err := s.canal.RunFrom(startPos)
		done <- err
	}()

	// 等待上下文结束或canal出错
	select {
	case <-ctx.Done():
		// 上下文取消，停止canal
		s.canal.Close()
		return ctx.Err()
	case err := <-done:
		// canal出错，更新运行状态
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("canal operation error: %w", err)
	}
}

// GetChanEventData 返回事件通道，供外部读取
// 返回值: 只读的事件通道，外部通过此通道接收数据库变更事件
func (s *MySQLSource) GetChanEventData() <-chan types.EventData {
	return s.eventDataChan
}

// Close 关闭数据源，释放资源
// 返回值: 可能的错误
func (s *MySQLSource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经停止运行，则直接返回
	if !s.running {
		return nil
	}

	// 关闭canal实例
	if s.canal != nil {
		s.canal.Close()
	}

	// 关闭事件通道
	close(s.eventDataChan)
	s.running = false
	return nil
}

// OnRow 处理行变更事件（实现canal.EventHandler接口）
// e: 行变更事件对象
// 返回值: 可能的错误
func (s *MySQLSource) OnRow(rowsEvent *canal.RowsEvent) error {
	// 获取当前时间戳（毫秒）
	// 处理每一行数据
	for i := 0; i < len(rowsEvent.Rows); i++ {
		row := rowsEvent.Rows[i]
		var event types.EventData
		event.Time = time.Now()
		event.Pos = int64(rowsEvent.Header.LogPos)
		event.ServerID = int64(rowsEvent.Header.ServerID)
		// 填充事件基本信息
		event.Row.Time = int64(rowsEvent.Header.Timestamp)
		event.Row.Database = rowsEvent.Table.Schema
		event.Row.Table = rowsEvent.Table.Name
		// 根据操作类型处理不同的事件
		switch rowsEvent.Action {
		case canal.InsertAction:
			event.Row.Type = types.InsertEventRowType
			event.Row.Data = s.rowToMap(row, rowsEvent.Table)
		case canal.DeleteAction:
			event.Row.Type = types.DeleteEventRowType
			event.Row.Data = s.rowToMap(row, rowsEvent.Table)
		case canal.UpdateAction:
			event.Row.Type = types.UpdateEventRowType
			oldRow := row
			// 更新操作会有两行数据：旧数据和新数据
			if i+1 < len(rowsEvent.Rows) {
				newRow := rowsEvent.Rows[i+1]
				event.Row.Data = s.rowToMap(newRow, rowsEvent.Table)
				event.Row.Old = s.rowToMap(oldRow, rowsEvent.Table)
				i++ // 跳过下一行（新数据）
			} else {
				event.Row.Data = s.rowToMap(row, rowsEvent.Table)
			}
		default:
			event.Row.Type = types.EventRowType(rowsEvent.Action)
			event.Row.Data = s.rowToMap(row, rowsEvent.Table)
		}
		select {
		case s.eventDataChan <- event:
		default:
			slog.Warn("Event channel is full, discarding event", "db", event.Row.Database, "table", event.Row.Table)
		}
	}

	return nil
}

// OnPosSynced 处理binlog位置同步事件（实现canal.EventHandler接口）
// header: 事件头部信息
// pos: 当前同步位置
// set: GTID集合
// force: 是否强制同步
// 返回值: 可能的错误
func (s *MySQLSource) OnPosSynced(header *replication.EventHeader, pos mysql.Position, set mysql.GTIDSet, force bool) error {
	// 保存当前同步位置
	return s.savePosition(pos)
}

// rowToMap 将数据库行数据转换为键值对映射
// row: 行数据数组
// table: 表结构信息
// 返回值: 字段名和值的映射
func (s *MySQLSource) rowToMap(row []interface{}, table *schema.Table) map[string]interface{} {
	m := make(map[string]interface{})

	for j, col := range table.Columns {
		raw := row[j]
		if raw == nil {
			m[col.Name] = nil
			continue
		}

		switch col.Type {
		case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT:
			switch v := raw.(type) {
			case []byte:
				val, err := strconv.ParseInt(string(v), 10, 64)
				if err != nil {
					m[col.Name] = v
				} else {
					m[col.Name] = val
				}
			case int, int32, int64:
				m[col.Name] = v
			default:
				m[col.Name] = raw
			}

		case schema.TYPE_FLOAT, schema.TYPE_DECIMAL:
			switch v := raw.(type) {
			case []byte:
				val, err := strconv.ParseFloat(string(v), 64)
				if err != nil {
					m[col.Name] = v
				} else {
					m[col.Name] = val
				}
			case float32, float64:
				m[col.Name] = v
			default:
				m[col.Name] = raw
			}

		case schema.TYPE_BIT:
			switch v := raw.(type) {
			case []byte:
				if len(v) > 0 {
					m[col.Name] = v[0] != 0
				} else {
					m[col.Name] = false
				}
			default:
				m[col.Name] = raw
			}

		case schema.TYPE_DATETIME, schema.TYPE_TIMESTAMP, schema.TYPE_DATE, schema.TYPE_TIME:
			switch v := raw.(type) {
			case []byte:
				t, err := time.Parse("2006-01-02 15:04:05", string(v))
				if err != nil {
					m[col.Name] = string(v)
				} else {
					m[col.Name] = t
				}
			case time.Time:
				m[col.Name] = v
			default:
				m[col.Name] = raw
			}

		case schema.TYPE_STRING, schema.TYPE_ENUM, schema.TYPE_SET, schema.TYPE_JSON, schema.TYPE_BINARY, schema.TYPE_POINT:
			switch v := raw.(type) {
			case []byte:
				m[col.Name] = string(v)
			default:
				m[col.Name] = raw
			}

		default:
			m[col.Name] = raw
		}
	}

	return m
}

// loadPosition 加载上次保存的同步位置
// 返回值: MySQL binlog同步位置
func (s *MySQLSource) loadPosition() mysql.Position {
	// 尝试从存储中加载上次保存的位置
	positionBytes, err := s.store.Get(StoreKeyPosition)
	if err == nil && len(positionBytes) > 0 {
		var storePos MysqlPosition
		_ = json.Unmarshal(positionBytes, &storePos)
		if storePos.File != "" && storePos.Pos != 0 {
			// 设置位置信息
			return mysql.Position{Name: storePos.File, Pos: storePos.Pos}
		}
	}
	// 如果加载失败，尝试获取主库当前位置
	pos, err := s.canal.GetMasterPos()
	if err == nil {
		return pos
	}
	// 都失败则返回默认位置
	return mysql.Position{}
}

// savePosition 保存当前位置信息
// pos: 要保存的同步位置
// 返回值: 可能的错误
func (s *MySQLSource) savePosition(pos mysql.Position) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 转换位置格式
	positionBytes, err := json.Marshal(MysqlPosition{
		File: pos.Name,
		Pos:  pos.Pos,
	})
	if err != nil {
		return fmt.Errorf("marshal position error: %w", err)
	}
	return s.store.Set(StoreKeyPosition, positionBytes)
}

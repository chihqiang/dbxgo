package types

import "time"

// EventRowType 定义数据库操作类型枚举
// 表示捕获到的数据库变更事件类型

type EventRowType string

const (
	// InsertEventRowType 表示插入操作类型
	InsertEventRowType EventRowType = "insert"
	// UpdateEventRowType 表示更新操作类型
	UpdateEventRowType EventRowType = "update"
	// DeleteEventRowType 表示删除操作类型
	DeleteEventRowType EventRowType = "delete"
)

// EventData 是标准事件结构
// 表示从数据库捕获到的变更事件
type EventData struct {
	Time     time.Time    `json:"time"`
	ServerID int64        `json:"server_id"`
	Pos      int64        `json:"pos"`
	Row      EventRowData `json:"row"`
}

// EventRowData 是标准事件结构
// 表示从数据库捕获到的变更事件
type EventRowData struct {
	// Time 事件发生时间戳（毫秒）
	Time int64 `json:"time"`
	// Database 数据库名称
	Database string `json:"database"`
	// Table 数据表名称
	Table string `json:"table"`
	// Type 事件类型（insert/update/delete）
	Type EventRowType `json:"type"`
	// Data 新数据内容，以字段名和值的映射表示
	Data map[string]any `json:"data"`
	// Old 旧数据内容，仅在update事件中有值
	Old map[string]any `json:"old,omitempty"`
}

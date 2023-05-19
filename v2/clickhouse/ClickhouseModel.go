package clickhouse

import "time"

type LogVO struct {
	ServiceName string
	PodName     string
	ReqPath     string
	ReqMethod   string
	ReqProtocol string
	ResTime     uint64
	ReqTime     uint64
	ResStatus   uint32
	ResSize     uint64
	RemoteAddr  string
	RemotePort  uint32
	LocalAddr   string
	LocalPort   uint32
	CreatedAt   time.Time
	ReqHeaders  string
	Message     string
}

type LogForm struct {
	ReqTimeFrom  int64  `json:"reqTimeFrom"`
	ReqTimeTo    int64  `json:"reqTimeTo"`
	OrderByField string `json:"orderByField"`
	OrderByType  string `json:"orderByType"`
	CustomQuery  string `json:"customQuery"`
	LimitSize    int    `json:"limitSize"`
	LimitStart   int    `json:"limitStart"`
}

type SvcLogForm struct {
	ReqTimeFrom string `json:"reqTimeFrom"`
	ReqTimeTo   string `json:"reqTimeTo"`
	ServiceName string `json:"serviceName"`
	QueryWords  string `json:"queryWords"`
	LimitSize   int    `json:"limitSize"`
	LimitStart  int    `json:"limitStart"`
}

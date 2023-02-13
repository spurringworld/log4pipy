package main

import "time"

type Trafficlogs struct {
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
	Timestamp   time.Time
	ReqHeaders  string
	Message     string
}

type Message struct {
	Req struct {
		// Protocol string
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"req"`
	Res struct {
		Status int `json:"status"`
	} `json:"res"`

	ReqTime    int64  `json:"reqTime"`
	ReqSize    int    `json:"reqSize"`
	ResTime    int64  `json:"resTime"`
	ResSize    int    `json:"resSize"`
	EndTime    int64  `json:"endTime"`
	RemoteAddr string `json:"remoteAddr"`
	RemotePort int    `json:"remotePort"`
	LocalAddr  string `json:"localAddr"`
	LocalPort  int    `json:"localPort"`

	Service struct {
		Name        string `json:"name"`
		Target      string `json:"target"`
		IngressMode bool   `json:"ingressMode"`
	} `json:"service"`
	Pod struct {
		Ns   string `json:"ns"`
		IP   string `json:"ip"`
		Name string `json:"name"`
	} `json:"pod"`
	Node struct {
		IP   string `json:"ip"`
		Name string `json:"name"`
	} `json:"node"`

	Type        string `json:"type"`
	MeshName    string `json:"meshName"`
	ClusterName string `json:"clusterName"`
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

package main

type Trafficlogs struct {
	Message string `sql:"message"`

	// ReqSize     int    `sql:"reqSize"`
	// ResSize     int    `sql:"resSize"`
	// ReqTime     int64  `sql:"reqTime"`
	// ResTime     int64  `sql:"resTime"`
	// EndTime     int64  `sql:"endTime"`
	// ReqPath     string `sql:"reqPath"`
	// ReqMethod   string `sql:"reqMethod"`
	// ResStatus   int    `sql:"resStatus"`
	// RemoteAddr  string `sql:"remoteAddr"`
	// RemotePort  int    `sql:"remotePort"`
	// LocalAddr   string `sql:"localAddr"`
	// LocalPort   int    `sql:"localPort"`
	// BondType    string `sql:"bondType"`
	// ServiceName string `sql:"serviceName"`
	// MeshName    string `sql:"meshName"`
	// ClusterName string `sql:"clusterName"`
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
	ReqTimeStart int64  `json:"reqTimeStart"`
	ReqTimeEnd   int64  `json:"reqTimeEnd"`
	OrderByField string `json:"orderByField"`
	OrderByType  string `json:"orderByType"`
	CustomQuery  string `json:"customQuery"`
	LimitSize    int    `json:"limitSize"`
	LimitStart   int    `json:"limitStart"`
}

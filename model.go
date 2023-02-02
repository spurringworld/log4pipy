package main

type Trafficlogs struct {
	Message     string `sql:"message"`
	ReqSize     int    `sql:"req_size"`
	ResSize     int    `sql:"res_size"`
	ReqTime     int64  `sql:"req_time"`
	ResTime     int64  `sql:"res_time"`
	EndTime     int64  `sql:"end_time"`
	ReqPath     string `sql:"req_path"`
	ReqMethod   string `sql:"req_method"`
	ResStatus   int    `sql:"res_status"`
	RemoteAddr  string `sql:"remote_addr"`
	RemotePort  int    `sql:"remote_port"`
	LocalAddr   string `sql:"local_addr"`
	LocalPort   int    `sql:"local_port"`
	BondType    string `sql:"bond_type"`
	ServiceName string `sql:"service_name"`
	MeshName    string `sql:"mesh_name"`
	ClusterName string `sql:"cluster_name"`
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

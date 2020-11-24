package handler

type ClusterInfo struct {
	Name     string `json:"name"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Nodes    int    `json:"nodes"`
	Cpu      string `json:"cpu"`
	Ram      string `json:"ram"`
}

type ClustersRes struct {
	Clusters []ClusterInfo `json:"clusters"`
}

type NodeInfo struct {
	Name          string `json:"name"`
	Cluster       string `json:"cluster"`
	Status        string `json:"status"`
	Role          string `json:"role"`
	SystemVersion string `json:"system_version"`
	Cpu           string `json:"cpu"`
	Ram           string `json:"memory"`
	Pods          string `json:"pods"`
}

type NodeRes struct {
	Nodes []NodeInfo `json:"nodes"`
}

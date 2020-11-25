package handler

type ClustersRes struct {
	Clusters []ClusterInfo `json:"clusters"`
}

type ClusterInfo struct {
	Name     string `json:"name"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Nodes    int    `json:"nodes"`
	Cpu      string `json:"cpu"`
	Ram      string `json:"ram"`
}

type NodeRes struct {
	Nodes []NodeInfo `json:"nodes"`
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

type ProjectRes struct {
	Projects []ProjectInfo `json:"projects"`
}

type ProjectInfo struct {
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Cluster     string                 `json:"cluster"`
	CreatedTime string                 `json:"created_time"`
	Labels      map[string]interface{} `json:"labels"`
}

type PodRes struct {
	Pods []PodInfo `json:"pods"`
}

type PodInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Cluster     string `json:"cluster"`
	Project     string `json:"project"`
	PodIP       string `json:"pod_ip"`
	Node        string `json:"node"`
	NodeIP      string `json:"node_ip"`
	Cpu         string `json:"cpu"`
	Ram         string `json:"memory"`
	CreatedTime string `json:"created_time"`
}

type DeploymentRes struct {
	Deployments []DeploymentInfo `json:"deployments"`
}

type DeploymentInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Cluster     string `json:"cluster"`
	Project     string `json:"project"`
	PodIP       string `json:"pod_ip"`
	Node        string `json:"node"`
	NodeIP      string `json:"node_ip"`
	Cpu         string `json:"cpu"`
	Ram         string `json:"memory"`
	CreatedTime string `json:"created_time"`
}

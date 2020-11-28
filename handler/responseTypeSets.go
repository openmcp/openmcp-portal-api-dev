package handler

type ClustersRes struct {
	Clusters []ClusterInfo `json:"clusters"`
}

type ClusterInfo struct {
	Name     string `json:"name"`
	Region   string `json:"region"`
	Zones    string `json:"zone"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Nodes    int    `json:"nodes"`
	Cpu      string `json:"cpu"`
	Ram      string `json:"ram"`
	Disk     string `json:"disk"`
	Network  string `json:"network"`
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
	Image       string `json:"image"`
	CreatedTime string `json:"created_time"`
}

type ServicesRes struct {
	Services []ServicesInfo `json:"services"`
}

type ServicesInfo struct {
	Name       string `json:"name"`
	Cluster    string `json:"cluster"`
	Project    string `json:"project"`
	Type       string `json:"type"`
	Selector   string `json:"selector"`
	Port       string `json:"port"`
	CreateTime string `json:"created_time"`
}

type IngerssRes struct {
	Ingress []IngerssInfo `json:"ingress"`
}

type IngerssInfo struct {
	Name        string `json:"name"`
	Cluster     string `json:"cluster"`
	Project     string `json:"project"`
	Address     string `json:"address"`
	CreatedTime string `json:"created_time"`
}

type ClusterResourceUsage struct {
	Cpu     Unit `json:"cpu"`
	Memory  Unit `json:"memory"`
	Storage Unit `json:"storage"`
}

type Unit struct {
	Unit    string    `json:"unit"`
	NameVal []NameVal `json:"status"`
}

type NameVal struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type ClusterOverView struct {
	Info             BasicInfo            `json:"basic_info"`
	PusageTop5       ProjectUsageTop5     `json:"project_usage_top5"`
	NusageTop5       NodeUsageTop5        `json:"node_usage_top5"`
	CUsage           ClusterResourceUsage `json:"cluster_resource_usage"`
	KubernetesStatus []NameStatus         `json:"kubernetes_status"`
	Events           []Event              `json:"events"`
}

type ProjectUsageTop5 struct {
	CPU    PairList `json:"cpu"`
	Memory PairList `json:"memory"`
}

type NodeUsageTop5 struct {
	CPU    PairList `json:"cpu"`
	Memory PairList `json:"memory"`
}

type NameStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type BasicInfo struct {
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	KubernetesVersion string `json:"kubernetes_version"`
	Status            string `json:"status"`
}

type Event struct {
	Project string `json:"project"`
	Typenm  string `json:"type"`
	Reason  string `json:"reason"`
	Object  string `json:"object"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

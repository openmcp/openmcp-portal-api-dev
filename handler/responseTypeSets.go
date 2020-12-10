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

type NodeOverView struct {
	Info              NodeBasicInfo     `json:"basic_info"`
	KbNodeStatus      []NameStatus      `json:"kubernetes_node_status"`
	NodeResourceUsage NodeResourceUsage `json:"node_resource_usage"`
}

type NodeBasicInfo struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	Role            string `json:"role"`
	Kubernetes      string `json:"kubernetes"`
	KubernetesProxy string `json:"kubernetes_proxy"`
	IP              string `json:"ip"`
	OS              string `json:"os"`
	Docker          string `json:"docker"`
	CreatedTime     string `json:"created_time"`
	Taint           Taint  `json:"taint"`
}

type NodeResourceUsage struct {
	Cpu     Unit `json:"cpu"`
	Memory  Unit `json:"memory"`
	Storage Unit `json:"storage"`
	Pods    Unit `json:"pods"`
}

type Taint struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Taint string `json:"taint"`
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

type HPARes struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Cluster     string `json:"cluster"`
	Reference   string `json:"reference"`
	MinRepl     string `json:"min_repl"`
	MaxRepl     string `json:"max_repl"`
	CurrentRepl string `json:"current_repl"`
}

type VPARes struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Cluster    string `json:"cluster"`
	Reference  string `json:"reference"`
	UpdateMode string `json:"update_mode"`
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
	Services []ServiceInfo `json:"services"`
}

type ServiceInfo struct {
	Name        string `json:"name"`
	Cluster     string `json:"cluster"`
	Project     string `json:"project"`
	Type        string `json:"type"`
	Selector    string `json:"selector"`
	Port        string `json:"port"`
	CreatedTime string `json:"created_time"`
}

type ServiceOverview struct {
	Info   ServiceBasicInfo `json:"basic_info"`
	Pods   []PodInfo        `json:"pods"`
	Events []Event          `json:"events"`
}

type ServiceBasicInfo struct {
	Name            string `json:"name"`
	Project         string `json:"project"`
	Type            string `json:"type"`
	Cluster         string `json:"cluster"`
	ClusterIP       string `json:"cluster_ip"`
	ExternalIP      string `json:"external_ip"`
	SessionAffinity string `json:"session_affinity"`
	Selector        string `json:"selector"`
	Endpoints       string `json:"endpoints"`
	CreatedTime     string `json:"created_time"`
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

type VolumeRes struct {
	Volumes []VolumeInfo `json:"volumes"`
}

type VolumeInfo struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	Status      string `json:"status"`
	Capacity    string `json:"capacity"`
	CreatedTime string `json:"created_time"`
}

// Secret List
type SecretRes struct {
	Secrets []SecretInfo `json:"secrets"`
}
type SecretInfo struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	Type        string `json:"type"`
	CreatedTime string `json:"created_time"`
}

//Secret Overview
type SecretOverView struct {
	Info SecretInfo `json:"basic_info"`
	Data []Data     `json:"data"`
}

type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigmapRes struct {
	Configmaps []ConfigmapInfo `json:"configmaps"`
}

type ConfigmapInfo struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	Keys        string `json:"keys"`
	CreatedTime string `json:"created_time"`
}

//Secret Overview
type ConfigmapOverView struct {
	Info ConfigmapInfo `json:"basic_info"`
	Data []Data        `json:"data"`
}

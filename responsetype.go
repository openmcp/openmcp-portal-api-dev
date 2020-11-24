package main

type NameVal struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Attributes struct {
	Status string `json:"status"`
	// Attributes struct {
	// 	Status string `json:"status"`
	// } `json:"attributes"`
}
type ChildNode struct {
	Name       string     `json:"name"`
	Attributes Attributes `json:"attributes"`
}

type Region struct {
	Name       string      `json:"name"`
	Attributes Attributes  `json:"attributes"`
	Children   []ChildNode `json:"children"`
}

type DashboardRes struct {
	Clusters struct {
		ClustersCnt    int       `json:"counts"`
		ClustersStatus []NameVal `json:"status"`
	} `json:"clusters"`
	Nodes struct {
		NodesCnt    int       `json:"counts"`
		NodesStatus []NameVal `json:"status"`
	} `json:"nodes"`
	Pods struct {
		PodsCnt    int       `json:"counts"`
		PodsStatus []NameVal `json:"status"`
	} `json:"pods"`
	Projects struct {
		ProjectsCnt    int       `json:"counts"`
		ProjectsStatus []NameVal `json:"status"`
	} `json:"projects"`
	Regions []Region `json:"regions"`
}

package main

import (
	"net/http"
	"portal-api-server/handler"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"yamlapply",
		"POST",
		"/apis/yamlapply",
		YamlApply,
	},
	Route{
		"joinableclusters",
		"GET",
		"/apis/joinableclusters",
		GetJoinableClusters,
	},
	Route{
		"vpa",
		"GET",
		"/apis/vpa",
		GetVPALists,
	},

	Route{
		"hpa",
		"GET",
		"/apis/hpa",
		GetHPALists,
	},

	Route{
		"addeksnode",
		"POST",
		"/apis/addeksnode",
		AddEKSnode,
	},

	Route{
		"migration",
		"POST",
		"/apis/migration",
		Migration,
	},

	Route{
		"addec2node",
		"POST",
		"/apis/addec2node",
		Addec2node,
	},

	Route{
		"dashboard",
		"GET",
		"/apis/dashboard",
		Dashboard,
	},

	Route{
		"clusters",
		"GET",
		"/apis/clusters",
		handler.Clusters,
	},

	Route{
		"cluster-overview",
		"GET",
		"/apis/clusters/overview",
		handler.ClusterOverview,
	},

	Route{
		"nodes",
		"GET",
		"/apis/clusters/{clusterName}/nodes",
		handler.NodesInCluster,
	},

	Route{
		"nodes",
		"GET",
		"/apis/nodes",
		handler.Nodes,
	},

	Route{
		"projects",
		"GET",
		"/apis/projects",
		handler.Projects,
	},

	Route{
		"deployments",
		"GET",
		"/apis/deployments",
		handler.Deployments,
	},

	Route{
		"services",
		"GET",
		"/apis/services",
		handler.Services,
	},

	Route{
		"ingress",
		"GET",
		"/apis/ingress",
		handler.Ingress,
	},
	// Route{
	// 	"projects",
	// 	"GET",
	// 	"/apis/clsuters/{clusterName}/projects/{projectName}/deployments",
	// 	handler.Deployments,
	// },

	Route{
		"pods",
		"GET",
		"/apis/clusters/{clusterName}/pods",
		handler.PodsInCluster,
	},
	Route{
		"pod-overview",
		"GET",
		"/apis/pod/overview",
		handler.PodOverview,
	},

	Route{
		"pods",
		"GET",
		"/apis/pods",
		handler.Pods,
	},
}

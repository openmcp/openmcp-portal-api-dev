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
		"test",
		"GET",
		"/test",
		Test,
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
		"pods",
		"GET",
		"/apis/pods",
		handler.Pods,
	},

	Route{
		"WorkloadsDeploymentsOverviewList",
		"GET",
		"/api/v1/workload/deployments/overview/clusters",
		WorkloadsDeploymentsOverviewList,
	},

	Route{
		"WorkloadsPodsOverviewList",
		"GET",
		"/api/v1/clusters/{clusterName}/daemonsets/list",
		WorkloadsPodsOverviewList,
	},

	Route{
		"getDeploymentList",
		"GET",
		"/api/v1/clusters/{clusterName}/deployments/list",
		getDeploymentList,
	},

	Route{
		"getDeploymentDetail",
		"GET",
		"/api/v1/clusters/{clusterName}/namespaces/{namespaceName}/deployments/{deploymentName}/detail",
		getDeploymentDetail,
	},

	Route{
		"getDeploymentYaml",
		"GET",
		"/api/v1/clusters/{clusterName}/namespaces/{namespaceName}/deployments/{deploymentName}/yaml",
		getDeploymentYaml,
	},

	Route{
		"getDeploymentEvent",
		"GET",
		"/api/v1/clusters/{clusterName}/namespaces/{namespaceName}/deployments/{deploymentName}/events",
		getDeploymentEvent,
	},

	Route{
		"getClusterList",
		"GET",
		"/api/v1/clusters/list",
		getDeploymentEvent,
	},
}

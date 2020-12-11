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
		handler.GetJoinedClusters,
	},
	Route{
		"joinableclusters",
		"GET",
		"/apis/joinableclusters",
		handler.GetJoinableClusters,
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
		"node-overview",
		"GET",
		"/apis/nodes/{nodeName}",
		handler.NodeOverview,
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
		handler.GetDeployments,
	},
	Route{
		"deploymentsInProject",
		"GET",
		"/apis/clsuters/{clusterName}/projects/{projectName}/deployments",
		handler.GetDeploymentsInProject,
	},
	Route{
		"dns",
		"GET",
		"/apis/dns",
		handler.Dns,
	},

	Route{
		"services",
		"GET",
		"/apis/services",
		handler.Services,
	},

	Route{
		"servicesInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/services",
		handler.GetServicesInProject,
	},

	Route{
		"serviceOverview",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/services/{serviceName}",
		handler.GetServiceOverview,
	},

	Route{
		"ingress",
		"GET",
		"/apis/ingress",
		handler.Ingress,
	},

	Route{
		"ingressInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/ingress",
		handler.GetIngressInProject,
	},
	Route{
		"pod-overview",
		"GET",
		"/apis/pod/overview",
		handler.PodOverview,
	},

	Route{
		"ingressOverview",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/ingress/{ingressName}",
		handler.GetIngressOverview,
	},

	Route{
		"pods",
		"GET",
		"/apis/pods",
		handler.GetPods,
	},

	Route{
		"vpa",
		"GET",
		"/apis/vpa",
		handler.GetVPAs,
	},

	Route{
		"hpa",
		"GET",
		"/apis/hpa",
		handler.GetHPAs,
	},

	Route{
		"podsInCluster",
		"GET",
		"/apis/clusters/{clusterName}/pods",
		handler.GetPodsInCluster,
	},

	Route{
		"podsInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/pods",
		handler.GetPodsInProject,
	},

	Route{
		"pvcInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/volumes",
		handler.GetVolumes,
	},

	Route{
		"pvcOverview",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/volumes/{volumeName}",
		handler.GetVolumeOverview,
	},

	Route{
		"secretInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/secrets",
		handler.GetSecrets,
	},

	Route{
		"secretOverview",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/secrets/{secretName}",
		handler.GetSecretOverView,
	},

	Route{
		"configmapInProject",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/configmaps",
		handler.GetConfigmaps,
	},

	Route{
		"configmapOverview",
		"GET",
		"/apis/clusters/{clusterName}/projects/{projectName}/configmaps/{configmapName}",
		handler.GetConfigmapOverView,
	},
}

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*
1. get Deployments
// http://192.168.0.152:31635/apis/apps/v1/deployments?clustername=cluster1
name : items > metatdata > namespace
ready :
cluster : clsuterName
image : items > spec > containers > image
updatedtime

2. Find Deployments In projects(namespace)

*/

func Deployments(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	// vars := mux.Vars(r)
	// clusterName := vars["clusterName"]
	// projectName := vars["projectName"]

	// fmt.Println(clustrName, projectName)

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data

	resCluster := ClustersRes{}

	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		cluster := ClusterInfo{}
		region := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
		clustername := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

		statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
		statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
		statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)

		clusterStatus := "Healthy"
		if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
			clusterStatus = "Unhealthy"
		} else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
			clusterStatus = "Healthy"
		} else {
			clusterStatus = "Unknown"
		}
		cluster.Name = clustername
		cluster.Region = region
		cluster.Status = clusterStatus
		cluster.Provider = "-"

		// get node names, cpu(capacity)
		nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + clustername
		go CallAPI(token, nodeURL, ch)
		nodeResult := <-ch
		nodeData := nodeResult.data
		nodeItems := nodeData["items"].([]interface{})

		cpuCapSum := 0
		memoryCapSum := 0
		cpuUseSum := 0
		memoryUseSum := 0

		// get nodename, cpu capacity Information
		for _, element := range nodeItems {
			nodeName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			cpuCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)
			cpuCapInt, _ := strconv.Atoi(cpuCapacity)

			memoryCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
			memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
			memoryCapInt, _ := strconv.Atoi(memoryCapacity)

			cpuCapSum += cpuCapInt
			memoryCapSum += memoryCapInt

			clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clustername
			go CallAPI(token, clMetricURL, ch)
			clMetricResult := <-ch
			clMetricData := clMetricResult.data

			//  cluster CPU Usage, Memroy Usage 확인
			for _, element := range clMetricData["nodemetrics"].([]interface{}) {
				cpuUse := element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"].(string)
				cpuUse = strings.Split(cpuUse, "n")[0]
				cpuUseInt, _ := strconv.Atoi(cpuUse)

				memoryUse := element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"].(string)
				memoryUse = strings.Split(memoryUse, "Ki")[0]
				memoryUseInt, _ := strconv.Atoi(memoryUse)
				// fmt.Println("memoryUseInt", memoryUse, memoryUseInt)

				cpuUseSum += cpuUseInt
				memoryUseSum += memoryUseInt
			}
		}

		//calculate cpu, memory unit
		cpuUseSumF := float64(cpuUseSum) / 1000 / 1000 / 1000
		cpuUseSumS := fmt.Sprintf("%.1f", cpuUseSumF)
		memoryUseSumF := float64(memoryUseSum) / 1000 / 1000
		memoryUseSumS := fmt.Sprintf("%.1f", memoryUseSumF)
		memoryCapSumF := float64(memoryCapSum) / 1000 / 1000
		memoryCapSumS := fmt.Sprintf("%.1f", memoryCapSumF)

		cluster.Nodes = len(nodeItems)
		cluster.Cpu = cpuUseSumS + "/" + strconv.Itoa(cpuCapSum) + " Core"
		cluster.Ram = memoryUseSumS + "/" + memoryCapSumS + " Gi"
		resCluster.Clusters = append(resCluster.Clusters, cluster)
	}
	json.NewEncoder(w).Encode(resCluster.Clusters)
}

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

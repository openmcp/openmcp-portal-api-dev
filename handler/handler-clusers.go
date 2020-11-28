package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func Clusters(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data

	resCluster := ClustersRes{}

	//get clusters Information
	clusterNames := []string{}
	//set master cluster info
	clusterNames = append(clusterNames, "openmcp")
	cluster := ClusterInfo{}
	cluster.Name = "openmcp"
	cluster.Provider = "-"
	cluster.Zones = "KR"
	cluster.Region = "AS"
	resCluster.Clusters = append(resCluster.Clusters, cluster)

	for _, element := range clusterData["items"].([]interface{}) {
		cluster := ClusterInfo{}
		region := element.(map[string]interface{})["status"].(map[string]interface{})["region"].(string)
		zones := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

		// statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
		// statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
		// statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)

		// clusterStatus := "Healthy"
		// if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
		// 	clusterStatus = "Unhealthy"
		// } else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
		// 	clusterStatus = "Healthy"
		// } else {
		// 	clusterStatus = "Unknown"
		// }
		cluster.Name = clusterName
		cluster.Provider = "-"
		cluster.Zones = zones
		cluster.Region = region
		resCluster.Clusters = append(resCluster.Clusters, cluster)
		clusterNames = append(clusterNames, clusterName)
	}

	for i, cluster := range resCluster.Clusters {

		cluster.Status = "Healthy"

		// get node names, cpu(capacity)
		nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + cluster.Name
		go CallAPI(token, nodeURL, ch)
		nodeResult := <-ch
		nodeData := nodeResult.data
		nodeItems := nodeData["items"].([]interface{})

		cpuCapSum := 0
		memoryCapSum := 0
		fsCapSum := 0
		cpuUseSum := 0
		memoryUseSum := 0
		fsUseSum := 0
		networkSum := 0

		// get nodename, cpu capacity Information
		for _, element := range nodeItems {
			nodeName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			status := ""
			statusInfo := element.(map[string]interface{})["status"]
			var healthCheck = make(map[string]string)
			for _, elem := range statusInfo.(map[string]interface{})["conditions"].([]interface{}) {
				conType := elem.(map[string]interface{})["type"].(string)
				tf := elem.(map[string]interface{})["status"].(string)
				healthCheck[conType] = tf
			}

			if healthCheck["Ready"] == "True" && healthCheck["NetworkUnavailable"] == "False" && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
				// healthyNodeCnt++
				status = "Healthy"
			} else {
				if healthCheck["Ready"] == "Unknown" || healthCheck["NetworkUnavailable"] == "Unknown" || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
					status = "Unknown"
				} else {
					status = "Unhealthy"
				}
			}
			if status == "Healthy" {
				resCluster.Clusters[i].Status = "Healthy"
			} else {
				resCluster.Clusters[i].Status = "Unhealthy"
			}

			cpuCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)
			cpuCapInt, _ := strconv.Atoi(cpuCapacity)

			memoryCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
			memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
			memoryCapInt, _ := strconv.Atoi(memoryCapacity)

			cpuCapSum += cpuCapInt
			memoryCapSum += memoryCapInt

			clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + cluster.Name
			// fmt.Println("check usl ::: http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + cluster.Name)
			go CallAPI(token, clMetricURL, ch)
			clMetricResult := <-ch
			clMetricData := clMetricResult.data

			cpuUse := "0n"
			memoryUse := "0Ki"
			fsUse := "0Ki"
			fsCap := "0Ki"
			ntRx := "0"
			ntTx := "0"
			//  cluster CPU Usage, Memroy Usage 확인
			if clMetricData["nodemetrics"] != nil {
				for _, element := range clMetricData["nodemetrics"].([]interface{}) {

					cpuUseCheck := element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"]
					if cpuUseCheck == nil {
						cpuUse = "0n"
					} else {
						cpuUse = cpuUseCheck.(string)
					}
					cpuUse = strings.Split(cpuUse, "n")[0]
					cpuUseInt, _ := strconv.Atoi(cpuUse)

					memoryUseCheck := element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"]
					if memoryUseCheck == nil {
						memoryUse = "0Ki"
					} else {
						memoryUse = memoryUseCheck.(string)
					}
					memoryUse = strings.Split(memoryUse, "Ki")[0]
					memoryUseInt, _ := strconv.Atoi(memoryUse)

					cpuUseSum += cpuUseInt
					memoryUseSum += memoryUseInt

					fsCapCheck := element.(map[string]interface{})["fs"].(map[string]interface{})["FsCapacityBytes"]
					if fsCapCheck == nil {
						fsCap = "0Ki"
					} else {
						fsCap = fsCapCheck.(string)
					}
					fsCap = strings.Split(fsCap, "Ki")[0]
					fsCapInt, _ := strconv.Atoi(fsCap)
					fsCapSum += fsCapInt

					fsUseCheck := element.(map[string]interface{})["fs"].(map[string]interface{})["FsUsedBytes"]
					if fsUseCheck == nil {
						fsUse = "0Ki"
					} else {
						fsUse = fsUseCheck.(string)
					}
					fsUse = strings.Split(fsUse, "Ki")[0]
					fsUseInt, _ := strconv.Atoi(fsUse)
					fsUseSum += fsUseInt

					ntRxCheck := element.(map[string]interface{})["network"].(map[string]interface{})["NetworkRxBytes"]
					if ntRxCheck == nil {
						ntRx = "0"
					} else {
						ntRx = ntRxCheck.(string)
					}
					ntTxCheck := element.(map[string]interface{})["network"].(map[string]interface{})["NetworkTxBytes"]
					if ntTxCheck == nil {
						ntTx = "0"
					} else {
						ntTx = ntTxCheck.(string)
					}
					ntTxUseInt, _ := strconv.Atoi(ntTx)
					ntRxUseInt, _ := strconv.Atoi(ntRx)
					rTxSum := ntRxUseInt + ntTxUseInt

					networkSum += rTxSum

				}
			}
		}

		//calculate cpu, memory unit
		cpuUseSumF := float64(cpuUseSum) / 1000 / 1000 / 1000
		cpuUseSumS := fmt.Sprintf("%.1f", cpuUseSumF)
		memoryUseSumF := float64(memoryUseSum) / 1000 / 1000
		memoryUseSumS := fmt.Sprintf("%.1f", memoryUseSumF)
		memoryCapSumF := float64(memoryCapSum) / 1000 / 1000
		memoryCapSumS := fmt.Sprintf("%.1f", memoryCapSumF)

		fsUseSumF := float64(fsUseSum) / 1000 / 1000
		fsUseSumS := fmt.Sprintf("%.1f", fsUseSumF)
		fsCapSumF := float64(fsCapSum) / 1000 / 1000
		fsCapSumS := fmt.Sprintf("%.1f", fsCapSumF)
		networkSumS := strconv.Itoa(networkSum)

		// fmt.Println(fsUseSumS, fsCapSumS)

		resCluster.Clusters[i].Nodes = len(nodeItems)
		resCluster.Clusters[i].Cpu = cpuUseSumS + "/" + strconv.Itoa(cpuCapSum) + " Core"
		resCluster.Clusters[i].Ram = memoryUseSumS + "/" + memoryCapSumS + " Gi"
		resCluster.Clusters[i].Disk = PercentUseString(fsUseSumS, fsCapSumS) + "%"
		resCluster.Clusters[i].Network = networkSumS + " bytes"
	}
	fmt.Println(resCluster.Clusters)
	json.NewEncoder(w).Encode(resCluster.Clusters)
}

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

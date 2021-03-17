package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type NodeClusters struct {
	Clusters []NodeCluster `json:"clusters"`
}

type NodeCluster struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Zones  string `json:"zone"`
}

func Nodes(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterURL := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterURL, ch)
	clusters := <-ch
	clusterData := clusters.data

	resNode := NodeRes{}
	podsCount := make(map[string]int)
	provider := "-"

	nodeClusters := NodeClusters{}
	nodeCluster := NodeCluster{}

	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")

	nodeCluster.Name = "openmcp"
	nodeCluster.Zones = "KR"
	nodeCluster.Region = "AS"
	nodeClusters.Clusters = append(nodeClusters.Clusters, nodeCluster)

	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		// provider := GetStringElement(element, []string{"metadata", "provider"})

		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		clusterRegion := GetStringElement(element, []string{"status", "region"})
		clusterZones := GetStringElement(element, []string{"status", "zones"})

		if clusterType == "Ready" {
			nodeCluster.Name = clusterName
			nodeCluster.Region = clusterRegion
			nodeCluster.Zones = clusterZones

			nodeClusters.Clusters = append(nodeClusters.Clusters, nodeCluster)
			clusterNames = append(clusterNames, clusterName)
		}
	}

	// for _, clusterName := range clusterNames {
	for _, nodeCluster := range nodeClusters.Clusters {
		node := NodeInfo{}
		// get node names, cpu(capacity)
		// nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + clusterName
		nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + nodeCluster.Name
		go CallAPI(token, nodeURL, ch)
		nodeResult := <-ch
		nodeData := nodeResult.data
		nodeItems := nodeData["items"].([]interface{})

		// get nodename, cpu capacity Information
		for _, element := range nodeItems {

			nodeName := GetStringElement(element, []string{"metadata", "name"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			cpuCapacity := GetStringElement(element, []string{"status", "capacity", "cpu"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)

			memoryCapacity := GetStringElement(element, []string{"status", "capacity", "memory"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
			memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
			memoryCapInt, _ := strconv.Atoi(memoryCapacity)
			memoryUseFloat := float64(memoryCapInt) / 1000 / 1000
			memoryCapacity = fmt.Sprintf("%.1f", memoryUseFloat)

			podsCapacity := GetStringElement(element, []string{"status", "capacity", "pods"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["pods"].(string)

			status := ""
			statusInfo := GetInterfaceElement(element, []string{"status"})
			// element.(map[string]interface{})["status"]
			var healthCheck = make(map[string]string)
			for _, elem := range statusInfo.(map[string]interface{})["conditions"].([]interface{}) {
				conType := GetStringElement(elem, []string{"type"})
				// elem.(map[string]interface{})["type"].(string)
				tf := GetStringElement(elem, []string{"status"})
				// elem.(map[string]interface{})["status"].(string)
				healthCheck[conType] = tf
			}

			if healthCheck["Ready"] == "True" && (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "False") && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
				// healthyNodeCnt++
				status = "Healthy"
			} else {
				if healthCheck["Ready"] == "Unknown" || (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "Unknown") || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
					status = "Unknown"
				} else {
					status = "Unhealthy"
				}
			}

			//정보유무 체크해야함
			role := ""
			roleCheck := GetInterfaceElement(element, []string{"metadata", "labels", "node-role.kubernetes.io/master"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["node-role.kubernetes.io/master"]

			if roleCheck == "" {
				role = "master"
			} else {
				role = "worker"
			}

			os := GetStringElement(element, []string{"status", "nodeInfo", "osImage"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["osImage"].(string)

			containerRuntimeVersion := GetStringElement(element, []string{"status", "nodeInfo", "containerRuntimeVersion"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["containerRuntimeVersion"].(string)

			clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + nodeCluster.Name
			// clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName
			// clMetricURL := "http://192.168.0.152:31635/metrics/nodes/clusterd-worker1.dev.gmd.life?clustername=cluster2"

			// fmt.Println("check usl ::: http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName)
			go CallAPI(token, clMetricURL, ch)
			clMetricResult := <-ch
			clMetricData := clMetricResult.data

			cpuUse := "0"
			memoryUse := "0"
			//  cluster CPU Usage, Memroy Usage 확인
			if clMetricData["nodemetrics"] != nil {
				for _, element := range clMetricData["nodemetrics"].([]interface{}) {
					cpuUseCheck := GetInterfaceElement(element, []string{"cpu", "CPUUsageNanoCores"})
					// element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"]
					if cpuUseCheck == nil {
						cpuUse = "0n"
					} else {
						cpuUse = cpuUseCheck.(string)
					}
					cpuUse = strings.Split(cpuUse, "n")[0]
					cpuUseInt, _ := strconv.Atoi(cpuUse)
					cpuUseFloat := float64(cpuUseInt) / 1000 / 1000 / 1000
					cpuUse = fmt.Sprintf("%.1f", cpuUseFloat)

					memoryUseCheck := GetInterfaceElement(element, []string{"memory", "MemoryUsageBytes"})
					// element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"]
					if memoryUseCheck == nil {
						memoryUse = "0Ki"
					} else {
						memoryUse = memoryUseCheck.(string)
					}
					memoryUse = strings.Split(memoryUse, "Ki")[0]
					memoryUseInt, _ := strconv.Atoi(memoryUse)
					memoryUseFloat := float64(memoryUseInt) / 1000 / 1000
					memoryUse = fmt.Sprintf("%.1f", memoryUseFloat)
				}
			}

			// if 조건으로 테스트용 Provider 입력해보자
			if nodeCluster.Name == "cluster1" {
				provider = "eks"
			} else if nodeCluster.Name == "cluster2" {
				provider = "kvm"
			} else if nodeCluster.Name == "openmcp" {
				provider = "aks"
			}

			node.Name = nodeName
			node.Cluster = nodeCluster.Name
			// node.Cluster = clusterName
			node.Status = status
			node.Role = role
			node.SystemVersion = os + "|" + "(" + containerRuntimeVersion + ")"
			node.Cpu = PercentUseString(cpuUse, cpuCapacity) + "%" + "|" + cpuUse + " / " + cpuCapacity + " Core"
			node.Ram = PercentUseString(memoryUse, memoryCapacity) + "%" + "|" + memoryUse + " / " + memoryCapacity + " GIB"
			node.Pods = podsCapacity
			node.Provider = provider
			node.Region = nodeCluster.Region

			resNode.Nodes = append(resNode.Nodes, node)
		}

		//pods counts by nodename
		podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + nodeCluster.Name
		// podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
		go CallAPI(token, podURL, ch)
		podResult := <-ch
		podData := podResult.data
		podItems := podData["items"].([]interface{})
		// fmt.Println("podItmes len:", len(podItems))

		// get podUsage counts by nodename groups
		for _, element := range podItems {
			nodeCheck := GetInterfaceElement(element, []string{"spec", "nodeName"})
			//  element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"]
			nodeName := "-"
			if nodeCheck == nil {
				nodeName = "-"
				// fmt.Println(element.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
			} else {
				nodeName = nodeCheck.(string)
			}
			podsCount[nodeName]++
		}
	}

	// add podUsage information in Pods
	for key, _ := range podsCount {
		// fmt.Println(key, val)
		for i := range resNode.Nodes {
			if resNode.Nodes[i].Name == key {
				// a := podsCount[v.Name]
				podsUsage := strconv.Itoa(podsCount[resNode.Nodes[i].Name])
				capacity := resNode.Nodes[i].Pods
				resNode.Nodes[i].Pods = PercentUseString(podsUsage, capacity) + "%" + "|" + podsUsage + " / " + capacity + " pods"

				// fmt.Println(resNode.Nodes[i].Pods)
			}
		}
	}

	json.NewEncoder(w).Encode(resNode.Nodes)
}

func NodesInCluster(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	// fmt.Println(clusterName)

	resNode := NodeRes{}
	node := NodeInfo{}
	podsCount := make(map[string]int)

	// get node names, cpu(capacity)
	nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + clusterName
	go CallAPI(token, nodeURL, ch)
	nodeResult := <-ch
	nodeData := nodeResult.data
	nodeItems := nodeData["items"].([]interface{})

	// get nodename, cpu capacity Information
	for _, element := range nodeItems {
		nodeName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

		cpuCapacity := GetStringElement(element, []string{"status", "capacity", "cpu"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)

		memoryCapacity := GetStringElement(element, []string{"status", "capacity", "memory"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
		memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
		memoryCapInt, _ := strconv.Atoi(memoryCapacity)
		memoryUseFloat := float64(memoryCapInt) / 1000 / 1000
		memoryCapacity = fmt.Sprintf("%.1f", memoryUseFloat)

		podsCapacity := GetStringElement(element, []string{"status", "capacity", "pods"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["pods"].(string)

		status := ""
		statusInfo := GetInterfaceElement(element, []string{"status"})
		// element.(map[string]interface{})["status"]
		var healthCheck = make(map[string]string)
		for _, elem := range statusInfo.(map[string]interface{})["conditions"].([]interface{}) {
			conType := GetStringElement(elem, []string{"type"})
			// elem.(map[string]interface{})["type"].(string)
			tf := GetStringElement(elem, []string{"status"})
			// elem.(map[string]interface{})["status"].(string)
			healthCheck[conType] = tf
		}

		if healthCheck["Ready"] == "True" && (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "False") && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
			// healthyNodeCnt++
			status = "Healthy"
		} else {
			if healthCheck["Ready"] == "Unknown" || (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "Unknown") || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
				status = "Unknown"
			} else {
				status = "Unhealthy"
			}
		}

		//정보유무 체크해야함
		role := ""
		roleCheck := GetInterfaceElement(element, []string{"metadata", "labels", "node-role.kubernetes.io/master"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["node-role.kubernetes.io/master"]

		if roleCheck == "" {
			role = "master"
		} else {
			role = "worker"
		}

		os := GetStringElement(element, []string{"status", "nodeInfo", "osImage"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["osImage"].(string)

		containerRuntimeVersion := GetStringElement(element, []string{"status", "nodeInfo", "containerRuntimeVersion"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["containerRuntimeVersion"].(string)

		clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName

		// fmt.Println("check usl ::: http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName)

		go CallAPI(token, clMetricURL, ch)
		clMetricResult := <-ch
		clMetricData := clMetricResult.data

		cpuUse := "0"
		memoryUse := "0"
		//  cluster CPU Usage, Memroy Usage 확인
		if clMetricData["nodemetrics"] != nil {
			for _, element := range clMetricData["nodemetrics"].([]interface{}) {

				cpuUseCheck := GetInterfaceElement(element, []string{"cpu", "CPUUsageNanoCores"})
				// element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"]
				if cpuUseCheck == nil {
					cpuUse = "0n"
				} else {
					cpuUse = cpuUseCheck.(string)
				}

				cpuUse = strings.Split(cpuUse, "n")[0]
				cpuUseInt, _ := strconv.Atoi(cpuUse)
				cpuUseFloat := float64(cpuUseInt) / 1000 / 1000 / 1000
				cpuUse = fmt.Sprintf("%.1f", cpuUseFloat)

				memoryUseCheck := GetInterfaceElement(element, []string{"memory", "MemoryUsageBytes"})
				// element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"]
				if memoryUseCheck == nil {
					memoryUse = "0Ki"
				} else {
					memoryUse = memoryUseCheck.(string)
				}
				memoryUse = strings.Split(memoryUse, "Ki")[0]
				memoryUseInt, _ := strconv.Atoi(memoryUse)
				memoryUseFloat := float64(memoryUseInt) / 1000 / 1000
				memoryUse = fmt.Sprintf("%.1f", memoryUseFloat)
			}
		}

		node.Name = nodeName
		node.Cluster = clusterName
		node.Status = status
		node.Role = role
		node.SystemVersion = os + "|" + "(" + containerRuntimeVersion + ")"
		node.Cpu = PercentUseString(cpuUse, cpuCapacity) + "%" + "|" + cpuUse + " / " + cpuCapacity + " Core"
		node.Ram = PercentUseString(memoryUse, memoryCapacity) + "%" + "|" + memoryUse + " / " + memoryCapacity + " GIB"
		node.Pods = podsCapacity

		resNode.Nodes = append(resNode.Nodes, node)
	}

	//pods counts by nodename
	podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
	go CallAPI(token, podURL, ch)
	podResult := <-ch
	podData := podResult.data
	podItems := podData["items"].([]interface{})

	// get podUsage counts by nodename groups
	for _, element := range podItems {
		nodeCheck := GetInterfaceElement(element, []string{"spec", "nodeName"})
		// element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"]
		nodeName := "-"
		if nodeCheck == nil {
			nodeName = "-"
			// fmt.Println(element.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
		} else {
			nodeName = nodeCheck.(string)
		}
		podsCount[nodeName]++
	}

	// add podUsage information in Pods
	for key, _ := range podsCount {
		// fmt.Println(key, val)
		for i := range resNode.Nodes {
			if resNode.Nodes[i].Name == key {
				// a := podsCount[v.Name]
				podsUsage := strconv.Itoa(podsCount[resNode.Nodes[i].Name])
				capacity := resNode.Nodes[i].Pods
				resNode.Nodes[i].Pods = PercentUseString(podsUsage, capacity) + "%" + "|" + podsUsage + " / " + capacity + " pods"

				// fmt.Println(resNode.Nodes[i].Pods)
			}
		}
	}

	json.NewEncoder(w).Encode(resNode.Nodes)
}

func NodeOverview(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("clustername")

	vars := mux.Vars(r)
	nodeName := vars["nodeName"]

	if nodeName == "" {
		errorMSG := jsonErr{500, "failed", "need some params"}
		json.NewEncoder(w).Encode(errorMSG)
	} else {
		ch := make(chan Resultmap)
		token := GetOpenMCPToken()

		// http://192.168.0.152:31635/api/v1/nodes/cluster1-worker1.dev.gmd.life?clustername=cluster1
		nodeURL := "http://" + openmcpURL + "/api/v1/nodes/" + nodeName + "?clustername=" + clusterName
		go CallAPI(token, nodeURL, ch)

		nodeResult := <-ch
		nodeData := nodeResult.data

		nodeName := nodeName
		createdTime := GetStringElement(nodeData, []string{"metadata", "creationTimestamp"})
		role := "worker"
		roleCheck := GetInterfaceElement(nodeData, []string{"metadata", "labels", "node-role.kubernetes.io/master"})
		if roleCheck == "" {
			role = "master"
		}
		os := GetStringElement(nodeData, []string{"status", "nodeInfo", "osImage"})
		kebernetes := GetStringElement(nodeData, []string{"status", "nodeInfo", "kubeletVersion"})
		kubernetesProxy := GetStringElement(nodeData, []string{"status", "nodeInfo", "kubeProxyVersion"})
		docker := GetStringElement(nodeData, []string{"status", "nodeInfo", "containerRuntimeVersion"})
		ip := ""
		ipList := GetArrayElement(nodeData, []string{"status", "addresses"})
		if ipList != nil {
			for _, item := range ipList {
				if item.(map[string]interface{})["type"] == "InternalIP" {
					ip = item.(map[string]interface{})["address"].(string)
				}
			}
		}

		status := ""
		var kubeNodeStatus []NameStatus
		var healthCheck = make(map[string]string)
		for _, elem := range GetArrayElement(nodeData, []string{"status", "conditions"}) {
			conType := GetStringElement(elem, []string{"type"})
			conStatus := GetStringElement(elem, []string{"status"})
			healthCheck[conType] = conStatus

			name := ""
			kubeStatus := "Unhealthy"
			if conType == "Ready" {
				name = "Kubelet"
				if conStatus == "True" {
					kubeStatus = "Healthy"
				}
				kubeNodeStatus = append(kubeNodeStatus, NameStatus{name, kubeStatus})
			} else if conType == "PIDPressure" || conType == "DiskPressure" || conType == "MemoryPressure" {
				name = conType
				if conStatus == "False" {
					kubeStatus = "Healthy"
				}
				kubeNodeStatus = append(kubeNodeStatus, NameStatus{name, kubeStatus})
			}

		}

		if healthCheck["Ready"] == "True" && (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "False") && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
			// healthyNodeCnt++
			status = "Healthy"
		} else {
			if healthCheck["Ready"] == "Unknown" || (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "Unknown") || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
				status = "Unknown"
			} else {
				status = "Unhealthy"
			}
		}
		taint := Taint{"", "", ""}
		provider := r.URL.Query().Get("provider")
		basicInfo := NodeBasicInfo{nodeName, status, role, kebernetes, kubernetesProxy, ip, os, docker, createdTime, taint, provider, clusterName}

		// Node Resource Usage
		cpuCapacity := GetStringElement(nodeData, []string{"status", "capacity", "cpu"})
		cpuCapFloat, _ := strconv.ParseFloat(cpuCapacity, 64)
		memoryCapacity := GetStringElement(nodeData, []string{"status", "capacity", "memory"})
		memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
		memoryCapFloat, _ := strconv.ParseFloat(memoryCapacity, 64)

		clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName
		go CallAPI(token, clMetricURL, ch)
		clMetricResult := <-ch
		clMetricData := clMetricResult.data

		cpuUse := "0"
		memoryUse := "0"
		fsUse := "0"
		fsCapaUse := "0"
		cpuUseFloat := 0.0
		memoryUseFloat := 0.0
		fsUseFloat := 0.0
		fsCapaUseFloat := 0.0

		//  cluster CPU Usage, Memroy Usage 확인
		if clMetricData["nodemetrics"] != nil {
			for _, element := range clMetricData["nodemetrics"].([]interface{}) {
				cpuUseCheck := GetInterfaceElement(element, []string{"cpu", "CPUUsageNanoCores"})
				if cpuUseCheck == nil {
					cpuUse = "0n"
				} else {
					cpuUse = cpuUseCheck.(string)
				}
				cpuUse = strings.Split(cpuUse, "n")[0]
				cpuUseFloat, _ = strconv.ParseFloat(cpuUse, 64)

				memoryUseCheck := GetInterfaceElement(element, []string{"memory", "MemoryUsageBytes"})
				if memoryUseCheck == nil {
					memoryUse = "0Ki"
				} else {
					memoryUse = memoryUseCheck.(string)
				}
				memoryUse = strings.Split(memoryUse, "Ki")[0]
				memoryUseFloat, _ = strconv.ParseFloat(memoryUse, 64)

				fsUseCheck := GetInterfaceElement(element, []string{"fs", "FsUsedBytes"})
				if fsUseCheck == nil {
					fsUse = "0Ki"
				} else {
					fsUse = fsUseCheck.(string)
				}
				fsUse = strings.Split(fsUse, "Ki")[0]
				fsUseFloat, _ = strconv.ParseFloat(fsUse, 64)

				fsCapaCheck := GetInterfaceElement(element, []string{"fs", "FsCapacityBytes"})
				if fsCapaCheck == nil {
					fsCapaUse = "0Ki"
				} else {
					fsCapaUse = fsCapaCheck.(string)
				}
				fsCapaUse = strings.Split(fsCapaUse, "Ki")[0]
				fsCapaUseFloat, _ = strconv.ParseFloat(fsCapaUse, 64)
			}
		}

		//podUsage 확인
		//pods counts by nodename
		podsCapacity := GetStringElement(nodeData, []string{"status", "capacity", "pods"})
		podCapaUseCount, _ := strconv.ParseFloat(podsCapacity, 64)

		podUseCount := 0
		podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
		go CallAPI(token, podURL, ch)
		podResult := <-ch
		podData := podResult.data
		podItems := podData["items"].([]interface{})
		// fmt.Println("podItmes len:", len(podItems))

		// get podUsage counts by nodename groups
		for _, element := range podItems {
			nodeCheck := GetInterfaceElement(element, []string{"spec", "nodeName"})
			if nodeCheck.(string) == nodeName {
				podUseCount++
			}
		}

		var cpuStatus []NameVal
		var memStatus []NameVal
		var fsStatus []NameVal
		var podStatus []NameVal
		cpuStatus = append(cpuStatus, NameVal{"Used", math.Ceil(cpuUseFloat/1000/1000/1000*100) / 100})
		cpuStatus = append(cpuStatus, NameVal{"Total", cpuCapFloat})
		// cpuStatus = append(cpuStatus, NameVal{"Total", fmt.Sprintf("%.1f", float64(clusterCPUCapSum)/1000/1000/1000)})

		memStatus = append(memStatus, NameVal{"Used", math.Ceil(memoryUseFloat/1000/1000*100) / 100})
		memStatus = append(memStatus, NameVal{"Total", math.Ceil(float64(memoryCapFloat)/1000/1000*100) / 100})
		// memStatus = append(memStatus, NameVal{"Total", fmt.Sprintf("%.1f", float64(clusterMemoryCapSum)/1000/1000)})

		fsStatus = append(fsStatus, NameVal{"Used", math.Ceil(fsUseFloat/1000/1000*100) / 100})
		fsStatus = append(fsStatus, NameVal{"Total", math.Ceil(fsCapaUseFloat/1000/1000*100) / 100})

		podStatus = append(podStatus, NameVal{"Used", float64(podUseCount)})
		podStatus = append(podStatus, NameVal{"Total", math.Ceil(podCapaUseCount*100) / 100})

		cpuUnit := Unit{"core", cpuStatus}
		memUnit := Unit{"Gi", memStatus}
		fsUnit := Unit{"Gi", fsStatus}
		podUnit := Unit{"", podStatus}

		nodeResUsage := NodeResourceUsage{cpuUnit, memUnit, fsUnit, podUnit}

		responseJSON := NodeOverView{basicInfo, kubeNodeStatus, nodeResUsage}

		json.NewEncoder(w).Encode(responseJSON)

	}

}

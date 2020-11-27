package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func Nodes(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterURL := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterURL, ch)
	clusters := <-ch
	clusterData := clusters.data

	resNode := NodeRes{}
	podsCount := make(map[string]int)

	// clusterNames := []string
	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		node := NodeInfo{}
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		// clusterNames = append(clusterNames, clusterName)

		// get node names, cpu(capacity)
		nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + clusterName
		go CallAPI(token, nodeURL, ch)
		nodeResult := <-ch
		nodeData := nodeResult.data
		nodeItems := nodeData["items"].([]interface{})

		// get nodename, cpu capacity Information
		for _, element := range nodeItems {

			nodeName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			cpuCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)

			memoryCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
			memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
			memoryCapInt, _ := strconv.Atoi(memoryCapacity)
			memoryUseFloat := float64(memoryCapInt) / 1000 / 1000
			memoryCapacity = fmt.Sprintf("%.1f", memoryUseFloat)

			podsCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["pods"].(string)

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

			//정보유무 체크해야함
			role := ""
			roleCheck := element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["node-role.kubernetes.io/master"]

			if roleCheck == "" {
				role = "master"
			} else {
				role = "worker"
			}

			os := element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["osImage"].(string)

			containerRuntimeVersion := element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["containerRuntimeVersion"].(string)

			clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName
			go CallAPI(token, clMetricURL, ch)
			clMetricResult := <-ch
			clMetricData := clMetricResult.data

			cpuUse := ""
			memoryUse := ""
			//  cluster CPU Usage, Memroy Usage 확인
			for _, element := range clMetricData["nodemetrics"].([]interface{}) {
				cpuUse = element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"].(string)
				cpuUse = strings.Split(cpuUse, "n")[0]
				cpuUseInt, _ := strconv.Atoi(cpuUse)
				cpuUseFloat := float64(cpuUseInt) / 1000 / 1000 / 1000
				cpuUse = fmt.Sprintf("%.1f", cpuUseFloat)

				memoryUse = element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"].(string)
				memoryUse = strings.Split(memoryUse, "Ki")[0]
				memoryUseInt, _ := strconv.Atoi(memoryUse)
				memoryUseFloat := float64(memoryUseInt) / 1000 / 1000
				memoryUse = fmt.Sprintf("%.1f", memoryUseFloat)
			}

			node.Name = nodeName
			node.Cluster = clusterName
			node.Status = status
			node.Role = role
			node.SystemVersion = os + "|" + "(" + containerRuntimeVersion + ")"
			node.Cpu = PercentUseString(cpuUse, cpuCapacity) + "%" + "|" + cpuUse + " / " + cpuCapacity + "|Core"
			node.Ram = PercentUseString(memoryUse, memoryCapacity) + "%" + "|" + memoryUse + " / " + memoryCapacity + "|GIB"
			node.Pods = podsCapacity

			resNode.Nodes = append(resNode.Nodes, node)
		}

		//pods counts by nodename
		podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
		go CallAPI(token, podURL, ch)
		podResult := <-ch
		podData := podResult.data
		podItems := podData["items"].([]interface{})
		fmt.Println("podItmes len:", len(podItems))

		// get podUsage counts by nodename groups
		for _, element := range podItems {
			nodeCheck := element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"]
			nodeName := "-"
			if nodeCheck == nil {
				nodeName = "-"
				fmt.Println(element.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
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
				resNode.Nodes[i].Pods = PercentUseString(podsUsage, capacity) + "%" + "|" + podsUsage + " / " + capacity

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
	fmt.Println(clusterName)

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
		nodeName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

		cpuCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)

		memoryCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
		memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
		memoryCapInt, _ := strconv.Atoi(memoryCapacity)
		memoryUseFloat := float64(memoryCapInt) / 1000 / 1000
		memoryCapacity = fmt.Sprintf("%.1f", memoryUseFloat)

		podsCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["pods"].(string)

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

		//정보유무 체크해야함
		role := ""
		roleCheck := element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["node-role.kubernetes.io/master"]

		if roleCheck == "" {
			role = "master"
		} else {
			role = "worker"
		}

		os := element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["osImage"].(string)

		containerRuntimeVersion := element.(map[string]interface{})["status"].(map[string]interface{})["nodeInfo"].(map[string]interface{})["containerRuntimeVersion"].(string)

		clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterName
		go CallAPI(token, clMetricURL, ch)
		clMetricResult := <-ch
		clMetricData := clMetricResult.data

		cpuUse := ""
		memoryUse := ""
		//  cluster CPU Usage, Memroy Usage 확인
		for _, element := range clMetricData["nodemetrics"].([]interface{}) {
			cpuUse := "0n"
			cpuUseCheck := element.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"]
			if cpuUseCheck == nil {
				cpuUse = "0n"
			} else {
				cpuUse = cpuUseCheck.(string)
			}
			cpuUse = strings.Split(cpuUse, "n")[0]
			cpuUseInt, _ := strconv.Atoi(cpuUse)
			cpuUseFloat := float64(cpuUseInt) / 1000 / 1000 / 1000
			cpuUse = fmt.Sprintf("%.1f", cpuUseFloat)

			memoryUse := "0Ki"
			memoryUseCheck := element.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"]
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

		node.Name = nodeName
		node.Cluster = clusterName
		node.Status = status
		node.Role = role
		node.SystemVersion = os + "|" + "(" + containerRuntimeVersion + ")"
		node.Cpu = PercentUseString(cpuUse, cpuCapacity) + "%" + "|" + cpuUse + " / " + cpuCapacity + "|Core"
		node.Ram = PercentUseString(memoryUse, memoryCapacity) + "%" + "|" + memoryUse + " / " + memoryCapacity + "|GIB"
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
		nodeCheck := element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"]
		nodeName := "-"
		if nodeCheck == nil {
			nodeName = "-"
			fmt.Println(element.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
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
				resNode.Nodes[i].Pods = PercentUseString(podsUsage, capacity) + "%" + "|" + podsUsage + " / " + capacity

				// fmt.Println(resNode.Nodes[i].Pods)
			}
		}
	}

	json.NewEncoder(w).Encode(resNode.Nodes)
}

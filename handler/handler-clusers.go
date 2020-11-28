package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

func ClusterOverview(w http.ResponseWriter, r *http.Request) {
	clusterNm := r.URL.Query().Get("clustername")
	if clusterNm == "" {
		errorMSG := jsonErr{500, "failed", "need some params"}
		json.NewEncoder(w).Encode(errorMSG)
	} else {
		// GetInfluxNodesMetric()
		InitInfluxConfig()
		inf := NewInflux(InfluxConfig.Influx.Ip, InfluxConfig.Influx.Port, InfluxConfig.Influx.Username, InfluxConfig.Influx.Username)
		results := GetInfluxPodsMetric(clusterNm, inf)

		type podUsage struct {
			namespace string
			cpuUsage  int
			memUsage  int
		}
		namspaceUsage := make(map[string]podUsage)
		cpuSum := make(map[string]int)
		memSum := make(map[string]int)

		for _, result := range results {
			if result.Series != nil {
				for _, ser := range result.Series {
					namespace := ser.Tags["namespace"]
					// pod := ser.Tags["pod"]
					cpu := 0
					mem := 0
					if ser.Values[0][1] != nil {
						cpuUsage := ser.Values[0][1].(string)
						if cpuUsage != "0" {
							cpuUsage = cpuUsage[:len(cpuUsage)-1]
							cpu, _ = strconv.Atoi(cpuUsage)
						}
					}
					if ser.Values[0][6] != nil {
						memUsage := ser.Values[0][6].(string)
						if memUsage != "0" {
							memUsage = memUsage[:len(memUsage)-2]
							mem, _ = strconv.Atoi(memUsage)
						}
					}

					cpuSum[namespace] += cpu
					memSum[namespace] += mem
					namspaceUsage[namespace] = podUsage{namespace, cpuSum[namespace], memSum[namespace]}

				}
			}
		}

		ch := make(chan Resultmap)
		token := GetOpenMCPToken()

		nodeURL := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + clusterNm
		go CallAPI(token, nodeURL, ch)

		nodeResult := <-ch
		nodeData := nodeResult.data
		nodeItems := nodeData["items"].([]interface{})

		clusterCPUCapSum := 0
		clusterMemoryCapSum := 0

		nodeResCPU := make(map[string]float64)
		nodeResMem := make(map[string]float64)
		healthyNodeCnt := 0
		unknownNodeCnt := 0
		unhealthyNodeCnt := 0
		var nodeResCPUSum float64
		var nodeResMemSum float64
		var nodeResFSSum int
		var nodeResFSCapaSum int

		var nodeNameList []string
		var kubeVersion string
		for _, element := range nodeItems {

			status := element.(map[string]interface{})["status"]
			var healthCheck = make(map[string]string)
			kubeVersion = status.(map[string]interface{})["nodeInfo"].(map[string]interface{})["kubeletVersion"].(string)
			for _, elem := range status.(map[string]interface{})["conditions"].([]interface{}) {
				conType := elem.(map[string]interface{})["type"].(string)
				tf := elem.(map[string]interface{})["status"].(string)
				healthCheck[conType] = tf

			}

			if healthCheck["Ready"] == "True" && healthCheck["NetworkUnavailable"] == "False" && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
				healthyNodeCnt++
			} else {
				if healthCheck["Ready"] == "Unknown" || healthCheck["NetworkUnavailable"] == "Unknown" || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
					unknownNodeCnt++
				} else {
					unhealthyNodeCnt++
				}
			}

			nodeName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			cpuCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["cpu"].(string)
			cpuCapInt, _ := strconv.Atoi(cpuCapacity)
			nodeNameList = append(nodeNameList, nodeName)
			memoryCapacity := element.(map[string]interface{})["status"].(map[string]interface{})["capacity"].(map[string]interface{})["memory"].(string)
			memoryCapacity = strings.Split(memoryCapacity, "Ki")[0]
			memoryCapInt, _ := strconv.Atoi(memoryCapacity)

			clusterCPUCapSum += cpuCapInt
			clusterMemoryCapSum += memoryCapInt

			clMetricURL := "http://" + openmcpURL + "/metrics/nodes/" + nodeName + "?clustername=" + clusterNm
			go CallAPI(token, clMetricURL, ch)
			clMetricResult := <-ch
			clMetricData := clMetricResult.data

			cpuUse := "0n"
			memoryUse := "0Ki"
			fsUse := "0Ki"
			fsCapaUse := "0Ki"

			fmt.Println("clusterCPUCapSum", clusterCPUCapSum)
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

					fsUseCheck := element.(map[string]interface{})["fs"].(map[string]interface{})["FsUsedBytes"]
					if fsUseCheck == nil {
						fsUse = "0Ki"
					} else {
						fsUse = fsUseCheck.(string)
					}
					fsUse = strings.Split(fsUse, "Ki")[0]
					fsUseInt, _ := strconv.Atoi(fsUse)

					fsCapaCheck := element.(map[string]interface{})["fs"].(map[string]interface{})["FsCapacityBytes"]
					if fsCapaCheck == nil {
						fsCapaUse = "0Ki"
					} else {
						fsCapaUse = fsCapaCheck.(string)
					}
					fsCapaUse = strings.Split(fsCapaUse, "Ki")[0]
					fsCapaUseInt, _ := strconv.Atoi(fsCapaUse)

					nodeResCPU[nodeName] = float64(cpuUseInt)
					nodeResMem[nodeName] = float64(memoryUseInt) / 1000

					nodeResCPUSum += nodeResCPU[nodeName]
					nodeResMemSum += nodeResMem[nodeName]
					nodeResFSSum += fsUseInt
					nodeResFSCapaSum += fsCapaUseInt
				}
			}

		}

		clusterCPURes := make(map[string]float64)
		clusterMemoryRes := make(map[string]float64)
		clusterCPUCapSum = clusterCPUCapSum * 1000 * 1000 * 1000
		for _, res := range namspaceUsage {
			cpuval := PercentChange(float64(res.cpuUsage), float64(clusterCPUCapSum))
			clusterCPURes[res.namespace] = math.Ceil(cpuval*100) / 100
			clusterMemoryRes[res.namespace] = float64(res.memUsage / 1000)
		}

		clusterCPURank := reverseRank(clusterCPURes, 5)
		clusterMemRank := reverseRank(clusterMemoryRes, 5)

		fmt.Println(clusterCPURank, clusterMemRank)

		for _, r := range nodeNameList {
			nodeCPUPecent := PercentChange(float64(nodeResCPU[r]), float64(clusterCPUCapSum))
			nodeResCPU[r] = math.Ceil(nodeCPUPecent*100) / 100
		}

		nodeCPURank := reverseRank(nodeResCPU, 5)
		nodeMemRank := reverseRank(nodeResMem, 5)
		fmt.Println(nodeCPURank, nodeMemRank)
		// nodeResCPUSumStr := fmt.Sprintf("%.1f", nodeResCPUSum/1000/1000/1000)
		nodeResCPUSumStr := nodeResCPUSum / 1000 / 1000 / 1000
		// nodeResMemSumStr := fmt.Sprintf("%.1f", nodeResMemSum/1000)
		nodeResMemSumStr := nodeResMemSum / 1000
		// nodeResFSSumStr := fmt.Sprintf("%.1f", float64(nodeResFSSum)/1000/1000)
		nodeResFSSumStr := float64(nodeResFSSum) / 1000 / 1000
		// nodeResFSCapaSumStr := fmt.Sprintf("%.1f", float64(nodeResFSCapaSum)/1000/1000)
		nodeResFSCapaSumStr := float64(nodeResFSCapaSum) / 1000 / 1000
		fmt.Println(nodeResCPUSumStr, nodeResMemSumStr, nodeResFSSum, nodeResFSCapaSum)

		config, _ := buildConfigFromFlags(clusterNm, kubeConfigFile)
		clientset, _ := kubernetes.NewForConfig(config)
		compStatus, _ := clientset.CoreV1().ComponentStatuses().List(v1.ListOptions{})

		nodeStatus := ""
		if unknownNodeCnt > 0 || unhealthyNodeCnt > 0 {
			nodeStatus = "Unhealthy"
		} else {
			nodeStatus = "Healthy"
		}

		var kubeStatus []NameStatus

		clusterStatus := "Healthy"
		for _, r := range compStatus.Items {
			name := r.Name
			conType := r.Conditions
			// status := r.Conditions[1]
			state := "Healthy"
			for _, c := range conType {
				if c.Type == "Healthy" && c.Status == "True" {
					state = "Healthy"
				} else {
					state = "Unhealthy"
					clusterStatus = "Unhealthy"
				}
			}

			if strings.Contains(name, "etcd") {
				name = "etcd"
			} else if strings.Contains(name, "scheduler") {
				name = "Scheduler"
			} else if strings.Contains(name, "controller-manager") {
				name = "Controller Manager"
			}

			kubeStatus = append(kubeStatus, NameStatus{name, state})

		}
		fmt.Println(kubeStatus)
		kubeStatus = append(kubeStatus, NameStatus{"Nodes", nodeStatus})

		eventsURL := "http://" + openmcpURL + "/api/v1/events?clustername=" + clusterNm
		go CallAPI(token, eventsURL, ch)
		eventsResult := <-ch
		eventsData := eventsResult.data
		eventsItems := eventsData["items"].([]interface{})
		var events []Event
		if eventsItems != nil {
			for _, element := range eventsItems {
				project := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
				typeNm := element.(map[string]interface{})["type"].(string)
				reason := element.(map[string]interface{})["reason"].(string)
				object := element.(map[string]interface{})["involvedObject"].(map[string]interface{})["kind"].(string)
				message := element.(map[string]interface{})["message"].(string)
				time := element.(map[string]interface{})["lastTimestamp"].(string)
				events = append(events, Event{project, typeNm, reason, object, message, time})
			}
		}

		var cpuStaus []NameVal
		var memStaus []NameVal
		var fsStaus []NameVal
		cpuStaus = append(cpuStaus, NameVal{"Used", math.Ceil(nodeResCPUSumStr*100) / 100})
		// cpuStaus = append(cpuStaus, NameVal{"Total", fmt.Sprintf("%.1f", float64(clusterCPUCapSum)/1000/1000/1000)})
		cpuStaus = append(cpuStaus, NameVal{"Total", math.Ceil(float64(clusterCPUCapSum)/1000/1000/1000*100) / 100})
		memStaus = append(memStaus, NameVal{"Used", math.Ceil(nodeResMemSumStr*100) / 100})
		// memStaus = append(memStaus, NameVal{"Total", fmt.Sprintf("%.1f", float64(clusterMemoryCapSum)/1000/1000)})
		memStaus = append(memStaus, NameVal{"Total", math.Ceil(float64(clusterMemoryCapSum)/1000/1000*100) / 100})
		fsStaus = append(fsStaus, NameVal{"Used", math.Ceil(nodeResFSSumStr*100) / 100})
		fsStaus = append(fsStaus, NameVal{"Total", math.Ceil(nodeResFSCapaSumStr*100) / 100})
		cpuUnit := Unit{"core", cpuStaus}
		memUnit := Unit{"Gi", memStaus}
		fsUnit := Unit{"Gi", fsStaus}
		cUsage := ClusterResourceUsage{cpuUnit, memUnit, fsUnit}
		info := BasicInfo{clusterNm, "-", kubeVersion, clusterStatus}

		pUsageTop5 := ProjectUsageTop5{clusterCPURank, clusterMemRank}
		nUsageTop5 := NodeUsageTop5{nodeCPURank, nodeMemRank}
		responseJSON := ClusterOverView{info, pUsageTop5, nUsageTop5, cUsage, kubeStatus, events}

		json.NewEncoder(w).Encode(responseJSON)

	}

}

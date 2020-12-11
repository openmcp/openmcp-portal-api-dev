package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func Pods(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterURL := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterURL, ch)
	clusters := <-ch
	clusterData := clusters.data

	resPod := PodRes{}
	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")
	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterNames = append(clusterNames, clusterName)
	}

	for _, clusterName := range clusterNames {
		podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
		go CallAPI(token, podURL, ch)
		podResult := <-ch
		podData := podResult.data
		podItems := podData["items"].([]interface{})

		// get podUsage counts by nodename groups
		for _, element := range podItems {
			pod := PodInfo{}
			podName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			project := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
			status := element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
			podIP := "-"
			node := "-"
			nodeIP := "-"
			if status == "Running" {
				podIP = element.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
				node = element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"].(string)
				nodeIP = element.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
			}

			cpu := "cpu"
			ram := "ram"
			createdTime := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

			pod.Name = podName
			pod.Status = status
			pod.Cluster = clusterName
			pod.Project = project
			pod.PodIP = podIP
			pod.Node = node
			pod.NodeIP = nodeIP
			pod.Cpu = cpu
			pod.Ram = ram
			pod.CreatedTime = createdTime

			resPod.Pods = append(resPod.Pods, pod)
		}
	}

	json.NewEncoder(w).Encode(resPod.Pods)
}

func PodsInCluster(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	resPod := PodRes{}

	podURL := "http://" + openmcpURL + "/api/v1/pods?clustername=" + clusterName
	go CallAPI(token, podURL, ch)
	podResult := <-ch
	podData := podResult.data
	podItems := podData["items"].([]interface{})

	// get podUsage counts by nodename groups
	for _, element := range podItems {
		pod := PodInfo{}
		podName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		project := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
		status := element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
		podIP := "-"
		node := "-"
		nodeIP := "-"
		if status == "Running" {
			podIP = element.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
			node = element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"].(string)
			nodeIP = element.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
		}
		cpu := "cpu"
		ram := "ram"
		createdTime := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

		pod.Name = podName
		pod.Status = status
		pod.Cluster = clusterName
		pod.Project = project
		pod.PodIP = podIP
		pod.Node = node
		pod.NodeIP = nodeIP
		pod.Cpu = cpu
		pod.Ram = ram
		pod.CreatedTime = createdTime

		resPod.Pods = append(resPod.Pods, pod)
	}

	json.NewEncoder(w).Encode(resPod.Pods)
}

func PodOverview(w http.ResponseWriter, r *http.Request) {
	clusterNm := r.URL.Query().Get("clustername")
	podNm := r.URL.Query().Get("pod")
	namespaceNm := r.URL.Query().Get("namespace")
	if clusterNm == "" || podNm == "" || namespaceNm == "" {
		errorMSG := jsonErr{500, "failed", "need some params"}
		json.NewEncoder(w).Encode(errorMSG)
	} else {
		ch := make(chan Resultmap)
		token := GetOpenMCPToken()
		// http://192.168.0.152:31635/api/v1/namespaces/{namespace}/pods/{podname}?clustername={clustername}

		podURL := "http://" + openmcpURL + "/api/v1/namespaces/" + namespaceNm + "/pods/" + podNm + "?clustername=" + clusterNm
		go CallAPI(token, podURL, ch)

		podResult := <-ch
		podData := podResult.data
		// fmt.Println(podData)
		if podData["spec"] != nil {
			podMetadata := podData["metadata"].(map[string]interface{})
			podSpec := podData["spec"].(map[string]interface{})
			podStatus := podData["status"].(map[string]interface{})
			totalRestartCount := 0
			var containers []PodOverviewContainer
			for _, element := range podStatus["containerStatuses"].([]interface{}) {
				restartCount := int(element.(map[string]interface{})["restartCount"].(float64))
				totalRestartCount = totalRestartCount + restartCount
			}
			for _, element := range podSpec["containers"].([]interface{}) {
				containerNm := element.(map[string]interface{})["name"].(string)
				containerImgae := element.(map[string]interface{})["image"].(string)
				containerPort := "-"
				if element.(map[string]interface{})["ports"] != nil {
					portInt := int(element.(map[string]interface{})["ports"].([]interface{})[0].(map[string]interface{})["containerPort"].(float64))
					containerPort = strconv.Itoa(portInt)
				} else {
					containerPort = "-"
				}
				restartCount := 0
				state := ""
				for _, contStatus := range podStatus["containerStatuses"].([]interface{}) {
					if contStatus.(map[string]interface{})["name"].(string) == containerNm {
						restartCount = int(contStatus.(map[string]interface{})["restartCount"].(float64))
						for k := range contStatus.(map[string]interface{})["state"].(map[string]interface{}) {
							state = k
						}
						break
					}
				}
				container := PodOverviewContainer{containerNm, state, restartCount, containerPort, containerImgae}
				containers = append(containers, container)
			}

			var podConditons []PodOverviewStatus

			for _, element := range podStatus["conditions"].([]interface{}) {
				conType := element.(map[string]interface{})["type"].(string)
				status := element.(map[string]interface{})["status"].(string)
				updateTime := element.(map[string]interface{})["lastTransitionTime"].(string)
				message := "-"
				reason := "-"
				if element.(map[string]interface{})["reason"] != nil {
					reason = element.(map[string]interface{})["reason"].(string)
				}
				if element.(map[string]interface{})["message"] != nil {
					message = element.(map[string]interface{})["message"].(string)
				}

				podConditons = append(podConditons, PodOverviewStatus{conType, status, updateTime, reason, message})
			}

			podBasicInfo := PodOverviewInfo{
				podMetadata["name"].(string),
				podStatus["phase"].(string),
				clusterNm,
				podMetadata["namespace"].(string),
				podStatus["podIP"].(string),
				podSpec["nodeName"].(string),
				podStatus["hostIP"].(string),
				podMetadata["namespace"].(string),
				strconv.Itoa(totalRestartCount),
				podMetadata["creationTimestamp"].(string),
			}

			podMetric := GetInfluxPod10mMetric(clusterNm, namespaceNm, podNm)

			// http://192.168.0.152:31635/api/v1/namespaces/{namespace}/events?clustername={clustername}
			podEventURL := "http://" + openmcpURL + "/api/v1/namespaces/" + namespaceNm + "/events?clustername=" + clusterNm
			go CallAPI(token, podEventURL, ch)

			eventsResult := <-ch
			eventsData := eventsResult.data
			eventsItems := eventsData["items"].([]interface{})
			var events []Event
			if eventsItems != nil {
				for _, element := range eventsItems {
					// project := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
					typeNm := element.(map[string]interface{})["type"].(string)
					reason := element.(map[string]interface{})["reason"].(string)
					objectKind := element.(map[string]interface{})["involvedObject"].(map[string]interface{})["kind"].(string)
					objectName := element.(map[string]interface{})["involvedObject"].(map[string]interface{})["name"].(string)
					message := element.(map[string]interface{})["message"].(string)
					time := "-"
					if element.(map[string]interface{})["lastTimestamp"] != nil {
						time = element.(map[string]interface{})["lastTimestamp"].(string)
					}

					if objectKind == "Pod" && objectName == podNm {
						events = append(events, Event{"", typeNm, reason, "", message, time})
					}
				}
			}

			response := PodOverviewRes{podBasicInfo, containers, podConditons, podMetric, events}

			json.NewEncoder(w).Encode(response)
		}
	}
}

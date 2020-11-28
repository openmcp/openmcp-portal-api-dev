package handler

import (
	"encoding/json"
	"net/http"

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

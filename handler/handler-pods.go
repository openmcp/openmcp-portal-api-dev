package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func GetPods(w http.ResponseWriter, r *http.Request) {
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
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		//  element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		if clusterType == "Ready" {
			clusterNames = append(clusterNames, clusterName)
		}
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
			podName := GetStringElement(element, []string{"metadata", "name"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			project := GetStringElement(element, []string{"metadata", "namespace"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
			status := GetStringElement(element, []string{"status", "phase"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
			podIP := "-"
			node := "-"
			nodeIP := "-"
			if status == "Running" {
				podIP = GetStringElement(element, []string{"status", "podIP"})
				// element.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
				node = GetStringElement(element, []string{"spec", "nodeName"})
				// element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"].(string)
				nodeIP = GetStringElement(element, []string{"status", "hostIP"})
				// element.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
			}

			cpu := "cpu"
			ram := "ram"
			createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

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

func GetPodsInCluster(w http.ResponseWriter, r *http.Request) {
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
		podName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		project := GetStringElement(element, []string{"metadata", "namespace"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
		status := GetStringElement(element, []string{"status", "phase"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
		podIP := "-"
		node := "-"
		nodeIP := "-"
		if status == "Running" {
			podIP = GetStringElement(element, []string{"status", "podIP"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
			node = GetStringElement(element, []string{"spec", "nodeName"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"].(string)
			nodeIP = GetStringElement(element, []string{"status", "hostIP"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
		}
		cpu := "cpu"
		ram := "ram"
		createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

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

func GetVPAs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	var allUrls []string

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data
	var clusternames []string
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusternames = append(clusternames, clusterName)
	}

	for _, cluster := range clusternames {
		vpaURL := "http://" + openmcpURL + "/apis/autoscaling.k8s.io/v1beta2/verticalpodautoscalers?clustername=" + cluster
		allUrls = append(allUrls, vpaURL)
	}

	for _, arg := range allUrls[0:] {
		go CallAPI(token, arg, ch)
	}

	var results = make(map[string]interface{})
	for range allUrls[0:] {
		result := <-ch
		results[result.url] = result.data
	}
	var VPAResList []VPARes

	for key, result := range results {
		clusterName := string(key[strings.LastIndex(key, "=")+1:])
		items := result.(map[string]interface{})["items"].([]interface{})
		for _, item := range items {
			hpaName := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			namespace := item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			reference := item.(map[string]interface{})["spec"].(map[string]interface{})["targetRef"].(map[string]interface{})["kind"].(string) + "/" + item.(map[string]interface{})["spec"].(map[string]interface{})["targetRef"].(map[string]interface{})["name"].(string)

			updateMode := item.(map[string]interface{})["spec"].(map[string]interface{})["updatePolicy"].(map[string]interface{})["updateMode"].(string)

			res := VPARes{hpaName, namespace, clusterName, reference, updateMode}

			VPAResList = append(VPAResList, res)

		}
	}
	json.NewEncoder(w).Encode(VPAResList)
}

func GetHPAs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	var allUrls []string

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data
	var clusternames []string
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusternames = append(clusternames, clusterName)
	}
	// add openmcp master cluster
	clusternames = append(clusternames, "openmcp")

	for _, cluster := range clusternames {
		hpaURL := "http://" + openmcpURL + "/apis/autoscaling/v1/horizontalpodautoscalers?clustername=" + cluster
		allUrls = append(allUrls, hpaURL)
	}

	for _, arg := range allUrls[0:] {
		go CallAPI(token, arg, ch)
	}

	var results = make(map[string]interface{})
	for range allUrls[0:] {
		result := <-ch
		results[result.url] = result.data
	}

	var HPAResList []HPARes

	for key, result := range results {
		clusterName := string(key[strings.LastIndex(key, "=")+1:])
		items := result.(map[string]interface{})["items"].([]interface{})
		for _, item := range items {
			hpaName := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			namespace := item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			reference := item.(map[string]interface{})["spec"].(map[string]interface{})["scaleTargetRef"].(map[string]interface{})["kind"].(string) + "/" + item.(map[string]interface{})["spec"].(map[string]interface{})["scaleTargetRef"].(map[string]interface{})["name"].(string)

			minRepl := item.(map[string]interface{})["spec"].(map[string]interface{})["minReplicas"].(float64)
			minReplStr := strconv.FormatFloat(minRepl, 'f', -1, 64)

			maxRepl := item.(map[string]interface{})["spec"].(map[string]interface{})["maxReplicas"].(float64)
			maxReplStr := strconv.FormatFloat(maxRepl, 'f', -1, 64)

			currentRepl := item.(map[string]interface{})["status"].(map[string]interface{})["currentReplicas"].(float64)
			currentRepllStr := strconv.FormatFloat(currentRepl, 'f', -1, 64)

			res := HPARes{hpaName, namespace, clusterName, reference, minReplStr, maxReplStr, currentRepllStr}

			HPAResList = append(HPAResList, res)

		}
	}
	json.NewEncoder(w).Encode(HPAResList)
}

func GetPodsInProject(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	projectName := vars["projectName"]
	resPod := PodRes{}

	// http: //192.168.0.152:31635/api/v1/namespaces/kube-system/pods?clustername=cluster2
	podURL := "http://" + openmcpURL + "/api/v1/namespaces/" + projectName + "/pods?clustername=" + clusterName
	go CallAPI(token, podURL, ch)
	podResult := <-ch
	podData := podResult.data
	podItems := podData["items"].([]interface{})

	// get podUsage counts by nodename groups
	for _, element := range podItems {
		pod := PodInfo{}
		podName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		project := GetStringElement(element, []string{"metadata", "namespace"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
		status := GetStringElement(element, []string{"status", "phase"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
		podIP := "-"
		node := "-"
		nodeIP := "-"
		if status == "Running" {
			podIP = GetStringElement(element, []string{"status", "podIP"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
			node = GetStringElement(element, []string{"spec", "nodeName"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["nodeName"].(string)
			nodeIP = GetStringElement(element, []string{"status", "hostIP"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
		}
		cpu := "cpu"
		ram := "ram"
		createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

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

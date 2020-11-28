package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

	resDeployment := DeploymentRes{}
	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")

	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterNames = append(clusterNames, clusterName)
	}

	for _, clusterName := range clusterNames {
		deployment := DeploymentInfo{}
		// get node names, cpu(capacity)
		deploymentURL := "http://" + openmcpURL + "/apis/apps/v1/deployments?clustername=" + clusterName
		go CallAPI(token, deploymentURL, ch)
		deploymentResult := <-ch
		// fmt.Println(deploymentResult)
		deploymentData := deploymentResult.data
		deploymentItems := deploymentData["items"].([]interface{})

		// get deployement Information
		for _, element := range deploymentItems {
			name := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			namespace := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			status := "-"
			availableReplicas := element.(map[string]interface{})["status"].(map[string]interface{})["availableReplicas"]
			readyReplicas := element.(map[string]interface{})["status"].(map[string]interface{})["readyReplicas"]
			replicas := element.(map[string]interface{})["status"].(map[string]interface{})["replicas"].(float64)

			replS := fmt.Sprintf("%.0f", replicas)

			if readyReplicas != nil {
				readyReplS := fmt.Sprintf("%.0f", readyReplicas)
				status = readyReplS + "/" + replS
			} else if availableReplicas == nil {
				status = "0/" + replS
			} else {
				status = "0/0"
			}

			image := element.(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"].(string)
			created_time := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

			deployment.Name = name
			deployment.Status = status
			deployment.Cluster = clusterName
			deployment.Project = namespace
			deployment.Image = image
			deployment.CreatedTime = created_time

			resDeployment.Deployments = append(resDeployment.Deployments, deployment)
		}
	}
	json.NewEncoder(w).Encode(resDeployment.Deployments)
}

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

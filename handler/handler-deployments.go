package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func GetDeployments(w http.ResponseWriter, r *http.Request) {
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
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		if clusterType == "Ready" {
			clusterNames = append(clusterNames, clusterName)
		}
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
			name := GetStringElement(element, []string{"metadata", "name"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			namespace := GetStringElement(element, []string{"metadata", "namespace"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			status := "-"
			availableReplicas := GetInterfaceElement(element, []string{"status", "availableReplicas"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["availableReplicas"]
			readyReplicas := GetInterfaceElement(element, []string{"status", "readyReplicas"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["readyReplicas"]
			replicas := GetFloat64Element(element, []string{"status", "replicas"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["replicas"].(float64)

			replS := fmt.Sprintf("%.0f", replicas)

			if readyReplicas != nil {
				readyReplS := fmt.Sprintf("%.0f", readyReplicas)
				status = readyReplS + "/" + replS
			} else if availableReplicas == nil {
				status = "0/" + replS
			} else {
				status = "0/0"
			}

			image := GetStringElement(element, []string{"spec", "template", "spec", "containers", "image"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"].(string)
			created_time := GetStringElement(element, []string{"metadata", "creationTimestamp"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

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

//get deployment-project list handler
func GetDeploymentsInProject(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	fmt.Println("GetDeploymentsInProject")

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	projectName := vars["projectName"]

	resDeployment := DeploymentRes{}
	deployment := DeploymentInfo{}
	// get node names, cpu(capacity)
	// http: //192.168.0.152:31635/apis/apps/v1/namespaces/kube-system/deployments?clustername=cluster1
	deploymentURL := "http://" + openmcpURL + "/apis/apps/v1/namespaces/" + projectName + "/deployments?clustername=" + clusterName
	go CallAPI(token, deploymentURL, ch)
	deploymentResult := <-ch
	// fmt.Println(deploymentResult)
	deploymentData := deploymentResult.data
	deploymentItems := deploymentData["items"].([]interface{})

	// get deployement Information
	for _, element := range deploymentItems {
		name := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		namespace := GetStringElement(element, []string{"metadata", "namespace"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

		status := "-"
		availableReplicas := GetInterfaceElement(element, []string{"status", "availableReplicas"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["availableReplicas"]
		readyReplicas := GetInterfaceElement(element, []string{"status", "readyReplicas"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["readyReplicas"]
		replicas := GetFloat64Element(element, []string{"status", "replicas"})
		// element.(map[string]interface{})["status"].(map[string]interface{})["replicas"].(float64)

		replS := fmt.Sprintf("%.0f", replicas)

		if readyReplicas != nil {
			readyReplS := fmt.Sprintf("%.0f", readyReplicas)
			status = readyReplS + "/" + replS
		} else if availableReplicas == nil {
			status = "0/" + replS
		} else {
			status = "0/0"
		}

		image := GetStringElement(element, []string{"spec", "template", "spec", "containers", "image"})
		// element.(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"].(string)
		created_time := GetStringElement(element, []string{"metadata", "creationTimestamp"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

		deployment.Name = name
		deployment.Status = status
		deployment.Cluster = clusterName
		deployment.Project = namespace
		deployment.Image = image
		deployment.CreatedTime = created_time

		resDeployment.Deployments = append(resDeployment.Deployments, deployment)
	}
	json.NewEncoder(w).Encode(resDeployment.Deployments)
}

//get cluster-node list handler

//get cluster-pods list handler

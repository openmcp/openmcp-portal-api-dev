package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Projects(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterURL := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterURL, ch)
	clusters := <-ch
	clusterData := clusters.data

	resProject := ProjectRes{}
	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")
	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterNames = append(clusterNames, clusterName)
	}

	for _, clusterName := range clusterNames {
		projectURL := "http://" + openmcpURL + "/api/v1/namespaces?clustername=" + clusterName
		go CallAPI(token, projectURL, ch)
		projectResult := <-ch
		projectData := projectResult.data
		projectItems := projectData["items"].([]interface{})

		// get podUsage counts by nodename groups
		for _, element := range projectItems {
			project := ProjectInfo{}
			projectName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			createdTime := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)
			status := element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)

			labels := make(map[string]interface{})
			labelCheck := element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"]
			if labelCheck == nil {
				//undefined lable
				labels = map[string]interface{}{}
			} else {
				for key, val := range labelCheck.(map[string]interface{}) {
					fmt.Println(key, val)
					labels[key] = val
				}
			}
			project.Name = projectName
			project.Status = status
			project.Cluster = clusterName
			project.CreatedTime = createdTime
			project.Labels = labels

			resProject.Projects = append(resProject.Projects, project)

		}
	}

	json.NewEncoder(w).Encode(resProject.Projects)
}
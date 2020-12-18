package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		if clusterType == "Ready" {
			clusterNames = append(clusterNames, clusterName)
		}
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
			projectName := GetStringElement(element, []string{"metadata", "name"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)
			status := GetStringElement(element, []string{"status", "phase"})
			// element.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)

			labels := make(map[string]interface{})
			labelCheck := GetInterfaceElement(element, []string{"metadata", "labels"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["labels"]
			if labelCheck == nil {
				//undefined lable
				labels = map[string]interface{}{}
			} else {
				for key, val := range labelCheck.(map[string]interface{}) {
					// fmt.Println(key, val)
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

func GetProjectOverview(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	projectName := vars["projectName"]

	resProjectOverview := ProjectOverview{}

	projectURL := "http://" + openmcpURL + "/api/v1/namespaces/" + projectName + "?clustername=" + clusterName
	go CallAPI(token, projectURL, ch)
	projectResult := <-ch
	projectData := projectResult.data
	// projectItems := projectData["items"].([]interface{})

	// get podUsage counts by nodename groups
	project := ProjectInfo{}
	createdTime := GetStringElement(projectData, []string{"metadata", "creationTimestamp"})
	status := GetStringElement(projectData, []string{"status", "phase"})
	uid := GetStringElement(projectData, []string{"metadata", "uid"})
	labels := make(map[string]interface{})
	labelCheck := GetInterfaceElement(projectData, []string{"metadata", "labels"})
	if labelCheck == nil {
		//undefined lable
		labels = map[string]interface{}{}
	} else {
		for key, val := range labelCheck.(map[string]interface{}) {
			labels[key] = val
		}
	}
	project.Name = projectName
	project.Status = status
	project.Cluster = clusterName
	project.CreatedTime = createdTime
	project.Labels = labels
	project.UID = uid
	resProjectOverview.Info = project

	// Project Resources
	// Pods, Deployments, StatefulSets, DaemonSets, Jobs (count, UnhealthyCount)

	// resProjectOverview.ProjectResource = append(resProjectOverview.ProjectResource, )

	// Usage Top5
	// workload (Deployment, Replicaset, statefulSet, DemonSet, Job, CronJob)
	// resProjectOverview.UsageTop5 = results

	results := GetInfluxDBPod10mMetric(clusterName, projectName)
	resProjectOverview.PhysicalResources = results

	json.NewEncoder(w).Encode(results)
}

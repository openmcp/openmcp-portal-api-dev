package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*
1. get services
http://192.168.0.152:31635/api/v1/services?clustername=cluster2
name : items > metatdata > name
namespace : items > metadata > namespace (project)
type : items > spec > type(string)
selector : items > spec > selector > 여러개 나옴 (key:value 형태로 가져오기)
port : items > spec > ports[여러개 나옴] > 안에 있는것 모두 나열
*/

func Services(w http.ResponseWriter, r *http.Request) {
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

	resServices := ServicesRes{}
	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")

	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterNames = append(clusterNames, clusterName)
	}

	for _, clusterName := range clusterNames {
		service := ServicesInfo{}
		serviceURL := "http://" + openmcpURL + "/api/v1/services?clustername=" + clusterName
		go CallAPI(token, serviceURL, ch)
		serviceResult := <-ch
		serviceData := serviceResult.data
		serviceItems := serviceData["items"].([]interface{})

		// get service Information
		for _, element := range serviceItems {
			name := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			namespace := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
			serviceType := element.(map[string]interface{})["spec"].(map[string]interface{})["type"].(string)

			selector := ""
			selectorCheck := element.(map[string]interface{})["spec"].(map[string]interface{})["selector"]
			if selectorCheck != nil {
				i := 0
				for key, val := range selectorCheck.(map[string]interface{}) {
					i++
					value := fmt.Sprintf("%v", val)
					if i == len(selectorCheck.(map[string]interface{})) {
						selector = selector + key + " : " + value
					} else {
						selector = selector + key + " : " + value + "|"
					}
				}
			} else {
				selector = "-"
			}

			port := ""
			portCheck := element.(map[string]interface{})["spec"].(map[string]interface{})["ports"].([]interface{})
			if portCheck != nil {
				for i, item := range portCheck {
					j := 0
					for key, val := range item.(map[string]interface{}) {
						j++
						value := fmt.Sprintf("%v", val)
						if j == len(item.(map[string]interface{})) {
							port = port + "{ " + key + " : " + value + " }"
						} else {
							port = port + "{ " + key + " : " + value + " },  "
						}
					}
					if i < len(portCheck)-1 {
						port = port + "|"
					}
				}

			} else {
				port = "-"
			}
			createdTime := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

			service.Cluster = clusterName
			service.Name = name
			service.Project = namespace
			service.Type = serviceType
			service.Selector = selector
			service.Port = port
			service.CreateTime = createdTime

			resServices.Services = append(resServices.Services, service)
		}
	}
	json.NewEncoder(w).Encode(resServices.Services)
}

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

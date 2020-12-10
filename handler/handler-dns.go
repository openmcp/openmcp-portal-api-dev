package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 1. get dns
// /apis/openmcp.k8s.io/v1alpha1/namespaces/default/openmcpdnsendpoints/
// name : items > metatdata > name
// namespace : items > metadata > namespace (project)
// type : items > spec > type(string)
// selector : items > spec > selector > 여러개 나옴 (key:value 형태로 가져오기)
// port : items > spec > ports[여러개 나옴] > 안에 있는것 모두 나열

func Dns(w http.ResponseWriter, r *http.Request) {
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
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		if clusterType == "Ready" {
			clusterNames = append(clusterNames, clusterName)
		}

	}

	for _, clusterName := range clusterNames {
		service := ServiceInfo{}

		dnsURL := "http://" + openmcpURL + "/apis/openmcp.k8s.io/v1alpha1/openmcpdnsendpoints?clustername=" + clusterName
		go CallAPI(token, dnsURL, ch)
		dnsResult := <-ch
		dnsData := dnsResult.data
		dnsItems := dnsData["items"].([]interface{})

		// get service Information
		for _, element := range dnsItems {
			name := GetStringElement(element, []string{"metadata", "name"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			namespace := GetStringElement(element, []string{"metadata", "namespace"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
			serviceType := GetStringElement(element, []string{"spec", "type"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["type"].(string)

			selector := ""
			selectorCheck := GetInterfaceElement(element, []string{"spec", "selector"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["selector"]
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
			portCheck := GetArrayElement(element, []string{"spec", "ports"})
			// element.(map[string]interface{})["spec"].(map[string]interface{})["ports"].([]interface{})
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
			createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})
			// element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

			service.Cluster = clusterName
			service.Name = name
			service.Project = namespace
			service.Type = serviceType
			service.Selector = selector
			service.Port = port
			service.CreatedTime = createdTime

			resServices.Services = append(resServices.Services, service)
		}
	}
	json.NewEncoder(w).Encode(resServices.Services)
}

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func Ingress(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data

	resIngress := IngerssRes{}
	clusterNames := []string{}
	clusterNames = append(clusterNames, "openmcp")

	//get clusters Information
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := GetStringElement(element, []string{"metadata", "name"})
		clusterType := GetStringElement(element, []string{"status", "conditions", "type"})
		if clusterType == "Ready" {
			clusterNames = append(clusterNames, clusterName)
		}
	}

	for _, clusterName := range clusterNames {
		ingress := IngerssInfo{}
		ingressURL := "http://" + openmcpURL + "/apis/networking.k8s.io/v1beta1/ingresses?clustername=" + clusterName
		go CallAPI(token, ingressURL, ch)

		ingressResult := <-ch
		ingressData := ingressResult.data
		ingressItems := ingressData["items"].([]interface{})

		if ingressItems != nil {
			for _, element := range ingressItems {
				name := GetStringElement(element, []string{"metadata", "name"})
				namespace := GetStringElement(element, []string{"metadata", "namespace"})
				ipAddress := GetStringElement(element, []string{"status", "loadBalancer", "ingress", "ip"})
				createTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})

				ingress.Name = name
				ingress.Project = namespace
				ingress.Address = ipAddress
				ingress.CreatedTime = createTime
				ingress.Cluster = clusterName

				resIngress.Ingress = append(resIngress.Ingress, ingress)
			}
		}
	}
	json.NewEncoder(w).Encode(resIngress.Ingress)
}

func GetIngressInProject(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	projectName := vars["projectName"]

	resIngress := IngerssRes{}
	ingress := IngerssInfo{}
	// "http://" + openmcpURL + "/api/v1/namespaces/" + projectName + "/services?
	ingressURL := "http://" + openmcpURL + "/apis/networking.k8s.io/v1beta1/namespaces/" + projectName + "/ingresses?clustername=" + clusterName

	go CallAPI(token, ingressURL, ch)

	ingressResult := <-ch
	ingressData := ingressResult.data
	ingressItems := ingressData["items"].([]interface{})

	if ingressItems != nil {
		for _, element := range ingressItems {
			name := GetStringElement(element, []string{"metadata", "name"})
			namespace := GetStringElement(element, []string{"metadata", "namespace"})
			ipAddress := GetStringElement(element, []string{"status", "loadBalancer", "ingress", "ip"})
			createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})

			ingress.Name = name
			ingress.Project = namespace
			ingress.Address = ipAddress
			ingress.CreatedTime = createdTime
			ingress.Cluster = clusterName

			resIngress.Ingress = append(resIngress.Ingress, ingress)
		}
	}
	json.NewEncoder(w).Encode(resIngress.Ingress)
}

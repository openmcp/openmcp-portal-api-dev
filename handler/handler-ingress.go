package handler

import (
	"encoding/json"
	"net/http"
)

/*
1. get services
http://192.168.0.152:31635/apis/networking.k8s.io/v1beta1/ingresses?clustername=cluster1
name : items > metatdata > name
namespace : items > metadata > namespace (project)
address : items > status > loadBalancer > ingress[] > ip
createdTime : items > metadata > creationTimestamp > 여러개 나옴 (key:value 형태로 가져오기)
*/

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
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusterNames = append(clusterNames, clusterName)
	}

	for _, clusterName := range clusterNames {
		ingress := IngerssInfo{}
		ingressURL := "http://" + openmcpURL + "/apis/networking.k8s.io/v1beta1/ingresses?clustername=" + clusterName
		go CallAPI(token, ingressURL, ch)
		ingressResult := <-ch
		ingressData := ingressResult.data
		ingressItems := ingressData["items"].([]interface{})

		// get service Information
		if ingressItems != nil {
			for _, element := range ingressItems {
				name := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
				namespace := element.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
				ipAddress := element.(map[string]interface{})["status"].(map[string]interface{})["loadBalancer"].(map[string]interface{})["ingress"].([]interface{})[0].(map[string]interface{})["ip"].(string)

				createTime := element.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)

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

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

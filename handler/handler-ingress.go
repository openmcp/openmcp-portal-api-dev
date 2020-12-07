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

// func findNestedMapValue(nMap interface{}, keys []string) interface{} {
// 	var result interface{}
// 	var childMap interface{}
// 	///"status", "loadBalancer", "ingress", "ip"
// 	for i, element := range keys {

// 		fmt.Println("value:", nMap.(map[string]interface{})[element])
// 		if nMap.(map[string]interface{})[element] != nil {
// 			childMap = nMap.(map[string]interface{})[element]
// 			i++
// 			keys = keys[i:]
// 			// fmt.Println()
// 			// fmt.Println("|||||", keys, len(keys))

// 			if len(keys) == 0 {
// 				result = nMap.(map[string]interface{})[element]
// 				// fmt.Println("break : ", result)
// 				break
// 			} else {
// 				// fmt.Println("findNestedMapValue()-----------", childMap, keys, reflect.TypeOf(childMap))
// 				typeCheck := fmt.Sprintf("%T", childMap)
// 				// fmt.Println(typeCheck)
// 				if "[]interface {}" == typeCheck {
// 					// typeCheck2 := fmt.Sprintf("%T", childMap[0].(string))
// 					// fmt.Println("true")
// 					result = findNestedMapValue(childMap.([]interface{})[0], keys)
// 					break
// 				} else {
// 					// fmt.Println("false")
// 					result = findNestedMapValue(childMap, keys)
// 					break
// 				}
// 			}
// 			// result = result + "/" + element + "(true)"
// 		} else {
// 			result = nil
// 			break
// 		}
// 	}
// 	return result
// }

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

//get cluster-overview list handler

//get cluster-node list handler

//get cluster-pods list handler

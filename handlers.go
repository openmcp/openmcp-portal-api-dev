package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"portal-api-server/cloud"
	"portal-api-server/resource"

	"github.com/gorilla/mux"
)

var targetURL = "172.17.1.241:7070"
var openmcpURL = "192.168.0.152:31635"

func Test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	node, ok1 := r.URL.Query()["node"]
	cluster, ok2 := r.URL.Query()["cluster"]
	if !ok1 || !ok2 || len(cluster[0]) < 1 || len(node[0]) < 1 {
		log.Println("Url Params are missing")
		w.Write([]byte("Url Param are missing"))
	} else {
		result := cloud.AddNode(node[0])
		if err := json.NewEncoder(w).Encode(result); err != nil {
			errmsg := jsonErr{444, "result fail"}
			json.NewEncoder(w).Encode(errmsg)
		}
		go cloud.GetNodeState(&result.InstanceID, node[0], cluster[0])

		// id := "i-09ce908be9488f77c"
		// cloud.GetNodeState(&id)
	}
}

func Clusters(w http.ResponseWriter, r *http.Request) {
	// // start := time.Now()
	// ch := make(chan Resultmap)
	// token := GetOpenMCPToken()

	// // var allUrls []string

	// clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	// go CallAPI(token, clusterurl, ch)
	// clusters := <-ch
	// // clusterData := clusters.data

	// fmt.Print("clusters:", clusters)
	// resCluster := DashboardRes{}
	// var clusterlist = make(map[string]Region)
	// var clusternames []string
	// clusterHealthyCnt := 0
	// clusterUnHealthyCnt := 0
	// clusterUnknownCnt := 0
	// for _, element := range clusterData["items"].([]interface{}) {
	// 	region := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
	// 	// if index > 0 {
	// 	// 	region = "US"
	// 	// }
	// 	clustername := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
	// 	statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
	// 	statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
	// 	statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)
	// 	clusterStatus := "Healthy"

	// 	if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
	// 		clusterStatus = "Unhealthy"
	// 		clusterUnHealthyCnt++
	// 	} else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
	// 		clusterStatus = "Healthy"
	// 		clusterHealthyCnt++
	// 	} else {
	// 		clusterStatus = "Unknown"
	// 		clusterUnknownCnt++
	// 	}
	// 	clusterlist[region] = Region{region, Attributes{clusterStatus}, append(clusterlist[region].Children, ChildNode{clustername, Attributes{clusterStatus}})}
	// 	clusternames = append(clusternames, clustername)
	// }

	// json.NewEncoder(w).Encode(resCluster)
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	// start := time.Now()
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	var allUrls []string

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data

	resCluster := DashboardRes{}
	var clusterlist = make(map[string]Region)
	var clusternames []string
	clusterHealthyCnt := 0
	clusterUnHealthyCnt := 0
	clusterUnknownCnt := 0
	for _, element := range clusterData["items"].([]interface{}) {
		region := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
		// if index > 0 {
		// 	region = "US"
		// }
		clustername := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
		statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
		statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)
		clusterStatus := "Healthy"

		if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
			clusterStatus = "Unhealthy"
			clusterUnHealthyCnt++
		} else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
			clusterStatus = "Healthy"
			clusterHealthyCnt++
		} else {
			clusterStatus = "Unknown"
			clusterUnknownCnt++
		}
		clusterlist[region] = Region{region, Attributes{clusterStatus}, append(clusterlist[region].Children, ChildNode{clustername, Attributes{clusterStatus}})}
		clusternames = append(clusternames, clustername)
	}

	for _, outp := range clusterlist {
		resCluster.Regions = append(resCluster.Regions, outp)
	}

	for _, cluster := range clusternames {
		nodeurl := "http://" + openmcpURL + "/api/v1/nodes?clustername=" + cluster
		allUrls = append(allUrls, nodeurl)
		podurl := "http://" + openmcpURL + "/api/v1/pods?clustername=" + cluster
		allUrls = append(allUrls, podurl)
		projecturl := "http://" + openmcpURL + "/api/v1/namespaces?clustername=" + cluster
		allUrls = append(allUrls, projecturl)
	}

	for _, arg := range allUrls[0:] {
		go CallAPI(token, arg, ch)
	}

	var results = make(map[string]interface{})
	nsCnt := 0
	podCnt := 0
	nodeCnt := 0

	for range allUrls[0:] {
		result := <-ch
		results[result.url] = result.data
	}

	ruuningPodCnt := 0
	failedPodCnt := 0
	unknownPodCnt := 0
	pendingPodCnt := 0
	activeNSCnt := 0
	terminatingNSCnt := 0
	healthyNodeCnt := 0
	unhealthyNodeCnt := 0
	unknownNodeCnt := 0

	for _, result := range results {

		kind := result.(map[string]interface{})["kind"]

		if kind == "NamespaceList" {
			nsCnt = nsCnt + len(result.(map[string]interface{})["items"].([]interface{}))
			for _, element := range result.(map[string]interface{})["items"].([]interface{}) {
				phase := element.(map[string]interface{})["status"].(map[string]interface{})["phase"]
				if phase == "Active" {
					activeNSCnt++
				} else if phase == "Terminating" {
					terminatingNSCnt++
				}
			}
		} else if kind == "PodList" {
			podCnt = podCnt + len(result.(map[string]interface{})["items"].([]interface{}))
			for _, element := range result.(map[string]interface{})["items"].([]interface{}) {
				phase := element.(map[string]interface{})["status"].(map[string]interface{})["phase"]
				if phase == "Running" {
					ruuningPodCnt++
				} else if phase == "Pending" {
					pendingPodCnt++
				} else if phase == "Failed" {
					failedPodCnt++
				} else if phase == "Unknown" {
					unknownPodCnt++
				}
			}

		} else if kind == "NodeList" {
			nodeCnt = nodeCnt + len(result.(map[string]interface{})["items"].([]interface{}))
			for _, element := range result.(map[string]interface{})["items"].([]interface{}) {
				status := element.(map[string]interface{})["status"]
				var healthCheck = make(map[string]string)
				for _, elem := range status.(map[string]interface{})["conditions"].([]interface{}) {
					conType := elem.(map[string]interface{})["type"].(string)
					tf := elem.(map[string]interface{})["status"].(string)
					healthCheck[conType] = tf

				}

				if healthCheck["Ready"] == "True" && healthCheck["NetworkUnavailable"] == "False" && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
					healthyNodeCnt++
				} else {
					if healthCheck["Ready"] == "Unknown" || healthCheck["NetworkUnavailable"] == "Unknown" || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
						unknownNodeCnt++
					} else {
						unhealthyNodeCnt++
					}
				}
			}
		}
	}

	resCluster.Clusters.ClustersCnt = len(clusternames)
	resCluster.Nodes.NodesCnt = nodeCnt
	resCluster.Pods.PodsCnt = podCnt
	resCluster.Projects.ProjectsCnt = nsCnt
	resCluster.Projects.ProjectsStatus = append(resCluster.Projects.ProjectsStatus, NameVal{"Active", activeNSCnt})
	resCluster.Projects.ProjectsStatus = append(resCluster.Projects.ProjectsStatus, NameVal{"Terminating", terminatingNSCnt})
	resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"Running", ruuningPodCnt})
	resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"Pending", pendingPodCnt})
	resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"Failed", failedPodCnt})
	resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"Unknown", unknownPodCnt})
	resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"Healthy", healthyNodeCnt})
	resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"Unhealthy", unhealthyNodeCnt})
	resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"Unknown", unknownNodeCnt})
	resCluster.Clusters.ClustersStatus = append(resCluster.Clusters.ClustersStatus, NameVal{"Healthy", clusterHealthyCnt})
	resCluster.Clusters.ClustersStatus = append(resCluster.Clusters.ClustersStatus, NameVal{"Unhealthy", clusterUnHealthyCnt})
	resCluster.Clusters.ClustersStatus = append(resCluster.Clusters.ClustersStatus, NameVal{"Unknown", clusterUnknownCnt})
	json.NewEncoder(w).Encode(resCluster)

	// nodeurl := "http://" + openmcpURL + "/api/v1/nodes?clustername=cluster1"
	// allUrls = append(allUrls, nodeurl)
	// podurl := "http://" + openmcpURL + "/api/v1/pods?clustername=cluster1"
	// allUrls = append(allUrls, podurl)
	// projecturl := "http://" + openmcpURL + "/api/v1/namespaces?clustername=cluster1"
	// allUrls = append(allUrls, projecturl)

	// for _, arg := range allUrls[0:] {
	// 	go CallAPI(token, arg, ch)

	// }

	// for _, arg := range allUrls[0:] {
	// 	// fmt.Println(<-ch)
	// 	results[arg] = <-ch
	// }

	// fmt.Println("%.2fs elapsed\n", time.Since(start).Seconds())
	// fmt.Println(results["http://192.168.0.152:31635/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"])
	// fmt.Println("%.2fs elapsed\n", time.Since(start).Seconds())
	// ******************************
	// if token != "" {

	// 	clusterData := CallAPI(token, clusterurl)
	// 	// nodeData := CallAPI(token, nodeurl)
	// 	// podData := CallAPI(token, podurl)
	// 	// projectData := CallAPI(token, projecturl)

	// 	fmt.Println("ClustersCnt:", len(clusterData["items"].([]interface{})))
	// 	// var children []ChildNode
	// 	// var childnode ChildNode
	// 	type zoneGroup struct {
	// 		zone       string
	// 		childNodes []ChildNode
	// 	}

	// 	resCluster := DashboardRes{}
	// 	var clusterlist = make(map[string]Region)

	// 	for _, element := range clusterData["items"].([]interface{}) {
	// 		region := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
	// 		// if index > 0 {
	// 		// 	region = "US"
	// 		// }
	// 		clustername := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
	// 		statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
	// 		statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
	// 		statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)
	// 		clusterStatus := "healthy"

	// 		if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
	// 			clusterStatus = "unhealthy"
	// 		} else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
	// 			clusterStatus = "healthy"
	// 		} else {
	// 			clusterStatus = "unknown"
	// 		}
	// 		clusterlist[region] = Region{region, Attributes{clusterStatus}, append(clusterlist[region].Children, ChildNode{clustername, Attributes{clusterStatus}})}

	// 		// fmt.Println("Index :", index, " Element :", reflect.TypeOf(element))

	// 		// clustername := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
	// 		// statusReason := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"].(string)
	// 		// statusType := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"].(string)
	// 		// statusTF := element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"].(string)
	// 		// region := element.(map[string]interface{})["status"].(map[string]interface{})["region"].(string)
	// 		// zones := element.(map[string]interface{})["status"].(map[string]interface{})["zones"].([]interface{})[0].(string)
	// 		// fmt.Println(region, "/", zones)
	// 		// if statusReason == "ClusterNotReachable" && statusType == "Offline" && statusTF == "True" {
	// 		// 	childnode = ChildNode{clustername, Attributes{"unhealthy"}}
	// 		// } else if statusReason == "ClusterReady" && statusType == "Ready" && statusTF == "True" {
	// 		// 	childnode = ChildNode{clustername, Attributes{"healthy"}}
	// 		// } else {
	// 		// 	childnode = ChildNode{clustername, Attributes{"unknown"}}
	// 		// }
	// 		// children = append(children, childnode)

	// 		// fmt.Println("Index :", index, "clustername : ", element.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
	// 		// fmt.Println("Index :", index, "status_reason : ", element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["reason"])
	// 		// fmt.Println("Index :", index, "status_type : ", element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["type"])
	// 		// fmt.Println("Index :", index, "status_tf : ", element.(map[string]interface{})["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})["status"])

	// 	}
	// 	// var outputs []zoneGroup

	// 	for _, outp := range clusterlist {
	// 		resCluster.Regions = append(resCluster.Regions, outp)
	// 	}

	// 	json.NewEncoder(w).Encode(resCluster)

	// 	// resCluster := DashboardRes{}
	// 	// resCluster.Clusters.ClustersCnt = 3
	// 	// resCluster.Clusters.ClustersStatus = append(resCluster.Clusters.ClustersStatus, NameVal{"healthy", 2})
	// 	// resCluster.Clusters.ClustersStatus = append(resCluster.Clusters.ClustersStatus, NameVal{"unhealthy", 1})
	// 	// resCluster.Nodes.NodesCnt = 10
	// 	// resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"healthy", 2})
	// 	// resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"unhealthy", 2})
	// 	// resCluster.Nodes.NodesStatus = append(resCluster.Nodes.NodesStatus, NameVal{"unknown", 2})
	// 	// resCluster.Pods.PodsCnt = 30
	// 	// resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"running", 10})
	// 	// resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"pending", 10})
	// 	// resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"unknown", 3})
	// 	// resCluster.Pods.PodsStatus = append(resCluster.Pods.PodsStatus, NameVal{"failed", 7})
	// 	// resCluster.Projects.ProjectsCnt = 1
	// 	// resCluster.Projects.ProjectsStatus = append(resCluster.Projects.ProjectsStatus, NameVal{"healthy", 13})
	// 	// resCluster.Projects.ProjectsStatus = append(resCluster.Projects.ProjectsStatus, NameVal{"unhealthy", 17})
	// 	// childnode := ChildNode{"C1", Attributes{"unhealthy"}}
	// 	// var children []ChildNode
	// 	// children = append(children, childnode)
	// 	// childnode = ChildNode{"C2", Attributes{"unhealthy"}}
	// 	// children = append(children, childnode)
	// 	// childnode = ChildNode{"C3", Attributes{"unhealthy"}}
	// 	// children = append(children, childnode)
	// 	// reg := Region{"aaaa", Attributes{"unhealthy"}, children}
	// 	// resCluster.Regions = append(resCluster.Regions, reg)
	// 	// children = nil
	// 	// children = append(children, childnode)
	// 	// childnode = ChildNode{"C4", Attributes{"healthy"}}
	// 	// children = append(children, childnode)
	// 	// childnode = ChildNode{"C5", Attributes{"healthy"}}
	// 	// children = append(children, childnode)
	// 	// reg = Region{"bbbb", Attributes{"healthy"}, children}
	// 	// resCluster.Regions = append(resCluster.Regions, reg)

	// 	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 	// w.WriteHeader(http.StatusOK)
	// 	// json.NewEncoder(w).Encode(resCluster)
	// 	// fmt.Println(resCluster)

	// 	// fmt.Println(data)

	// } else {
	// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte("Cannot Auth OpenMCP API Server"))
	// }
	// ******************************

	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte("clusters"))
}

func WorkloadsDeploymentsOverviewList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resource.ListResource()); err != nil {
		panic(err)
	}

}

func WorkloadsPodsOverviewList(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	vars := mux.Vars(r)

	var client http.Client
	resp, err := client.Get("https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/daemonsets/list")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

func getDeploymentList(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	vars := mux.Vars(r)

	var client http.Client
	resp, err := client.Get("https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/deployments/list")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

func getDeploymentDetail(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	vars := mux.Vars(r)

	var client http.Client

	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/detail"
	//fmt.Print(callUrl)

	resp, err := client.Get(callUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

func getDeploymentYaml(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	vars := mux.Vars(r)

	var client http.Client

	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/yaml"
	//fmt.Print(callUrl)

	resp, err := client.Get(callUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

func getDeploymentEvent(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	vars := mux.Vars(r)

	var client http.Client

	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/events"
	//fmt.Print(callUrl)

	resp, err := client.Get(callUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

package main

import (
	"encoding/json"
	"net/http"
	"portal-api-server/cloud"
	"portal-api-server/handler"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

var openmcpURL = handler.InitPortalConfig()

func GetJoinableClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	url := "http://" + openmcpURL + "/joinable"
	go CallAPI(token, url, ch)
	clusters := <-ch
	clusterData := clusters.data

	type joinable struct {
		Name     string `json:"name"`
		Endpoint string `json:"endpoint"`
		Platform string `json:"platform"`
		Region   string `json:"region"`
		Zone     string `json:"zone"`
	}

	var joinableLists []joinable
	if clusterData["items"] != nil {
		for _, element := range clusterData["items"].([]interface{}) {
			name := element.(map[string]interface{})["name"].(string)
			endpoint := element.(map[string]interface{})["endpoint"].(string)
			platform := element.(map[string]interface{})["platform"].(string)
			region := element.(map[string]interface{})["region"].(string)
			zone := element.(map[string]interface{})["zone"].(string)
			res := joinable{name, endpoint, platform, region, zone}
			joinableLists = append(joinableLists, res)
		}
		json.NewEncoder(w).Encode(joinableLists)
	} else {
		json.NewEncoder(w).Encode(joinableLists)
	}
}

func GetVPALists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	var allUrls []string

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data
	var clusternames []string
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusternames = append(clusternames, clusterName)
	}

	for _, cluster := range clusternames {
		vpaURL := "http://" + openmcpURL + "/apis/autoscaling.k8s.io/v1beta2/verticalpodautoscalers?clustername=" + cluster
		allUrls = append(allUrls, vpaURL)
	}

	for _, arg := range allUrls[0:] {
		go CallAPI(token, arg, ch)
	}

	var results = make(map[string]interface{})
	for range allUrls[0:] {
		result := <-ch
		results[result.url] = result.data
	}
	var VPAResList []VPARes

	for key, result := range results {
		clusterName := string(key[strings.LastIndex(key, "=")+1:])
		items := result.(map[string]interface{})["items"].([]interface{})
		for _, item := range items {
			hpaName := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			namespace := item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			reference := item.(map[string]interface{})["spec"].(map[string]interface{})["targetRef"].(map[string]interface{})["kind"].(string) + "/" + item.(map[string]interface{})["spec"].(map[string]interface{})["targetRef"].(map[string]interface{})["name"].(string)

			updateMode := item.(map[string]interface{})["spec"].(map[string]interface{})["updatePolicy"].(map[string]interface{})["updateMode"].(string)

			res := VPARes{hpaName, namespace, clusterName, reference, updateMode}

			VPAResList = append(VPAResList, res)

		}
	}
	json.NewEncoder(w).Encode(VPAResList)
}

func GetHPALists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	var allUrls []string

	clusterurl := "http://" + openmcpURL + "/apis/core.kubefed.io/v1beta1/kubefedclusters?clustername=openmcp"
	go CallAPI(token, clusterurl, ch)
	clusters := <-ch
	clusterData := clusters.data
	var clusternames []string
	for _, element := range clusterData["items"].([]interface{}) {
		clusterName := element.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		clusternames = append(clusternames, clusterName)
	}
	// add openmcp master cluster
	clusternames = append(clusternames, "openmcp")

	for _, cluster := range clusternames {
		hpaURL := "http://" + openmcpURL + "/apis/autoscaling/v1/horizontalpodautoscalers?clustername=" + cluster
		allUrls = append(allUrls, hpaURL)
	}

	for _, arg := range allUrls[0:] {
		go CallAPI(token, arg, ch)
	}

	var results = make(map[string]interface{})
	for range allUrls[0:] {
		result := <-ch
		results[result.url] = result.data
	}

	var HPAResList []HPARes

	for key, result := range results {
		clusterName := string(key[strings.LastIndex(key, "=")+1:])
		items := result.(map[string]interface{})["items"].([]interface{})
		for _, item := range items {
			hpaName := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)

			namespace := item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)

			reference := item.(map[string]interface{})["spec"].(map[string]interface{})["scaleTargetRef"].(map[string]interface{})["kind"].(string) + "/" + item.(map[string]interface{})["spec"].(map[string]interface{})["scaleTargetRef"].(map[string]interface{})["name"].(string)

			minRepl := item.(map[string]interface{})["spec"].(map[string]interface{})["minReplicas"].(float64)
			minReplStr := strconv.FormatFloat(minRepl, 'f', -1, 64)

			maxRepl := item.(map[string]interface{})["spec"].(map[string]interface{})["maxReplicas"].(float64)
			maxReplStr := strconv.FormatFloat(maxRepl, 'f', -1, 64)

			currentRepl := item.(map[string]interface{})["status"].(map[string]interface{})["currentReplicas"].(float64)
			currentRepllStr := strconv.FormatFloat(currentRepl, 'f', -1, 64)

			res := HPARes{hpaName, namespace, clusterName, reference, minReplStr, maxReplStr, currentRepllStr}

			HPAResList = append(HPAResList, res)

		}
	}
	json.NewEncoder(w).Encode(HPAResList)
}

func Migration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	clusterurl := "http://" + openmcpURL + "/apis/openmcp.k8s.io/v1alpha1/namespaces/default/migrations?clustername=openmcp"
	resp, err := PostYaml(clusterurl, r.Body)
	defer r.Body.Close()
	if err != nil {
		errmsg := jsonErr{503, "failed", "request fail"}
		json.NewEncoder(w).Encode(errmsg)
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(resp), &data)

	if data["kind"].(string) == "Status" {
		msg := jsonErr{501, "failed", data["message"].(string)}
		json.NewEncoder(w).Encode(msg)
	} else {
		msg := jsonErr{200, "success", "Migration Created"}
		json.NewEncoder(w).Encode(msg)
	}

}

func AddEKSnode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	akid := "AKIAJGFO6OXHRN2H6DSA"
	secretkey := "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"
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-2"),
		Credentials: credentials.NewStaticCredentials(akid, secretkey, ""),
	})

	if err != nil {
		errmsg := jsonErr{503, "failed", "result fail"}
		json.NewEncoder(w).Encode(errmsg)
	}

	svc := eks.New(sess)

	result, err := svc.ListNodegroups(&eks.ListNodegroupsInput{
		ClusterName: aws.String("testcluster"),
	})

	nodegroupname := result.Nodegroups[0]

	result2, err := svc.DescribeNodegroup(&eks.DescribeNodegroupInput{
		ClusterName:   aws.String("testcluster"),
		NodegroupName: aws.String(*nodegroupname),
	})

	beforecnt := result2.Nodegroup.ScalingConfig.DesiredSize
	var nodecnt int64
	nodecnt = -1
	desirecnt := *beforecnt + nodecnt

	// // la := make(map[string]*string)
	// // namelabel := "newlabel01"
	// // la["newlabel01"] = &namelabel

	// labelinput := eks.UpdateLabelsPayload{la["newlabel01"]}

	addResult, err := svc.UpdateNodegroupConfig(&eks.UpdateNodegroupConfigInput{
		ClusterName:   aws.String("testcluster"),
		NodegroupName: aws.String(*nodegroupname),
		// Labels:        &eks.UpdateLabelsPayload{AddOrUpdateLabels: la},
		ScalingConfig: &eks.NodegroupScalingConfig{DesiredSize: &desirecnt},
	})

	if err != nil {
		errmsg := jsonErr{503, "failed", "result fail"}
		json.NewEncoder(w).Encode(errmsg)
	}

	// fmt.Println(addResult)

}

func Addec2node(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	data := GetJsonBody(r.Body)
	defer r.Body.Close() // 리소스 누출 방지
	node := data["node"].(string)
	cluster := data["cluster"].(string)
	aKey := data["a_key"].(string)
	sKey := data["s_key"].(string)
	result := cloud.AddNode(node, aKey, sKey)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		errmsg := jsonErr{503, "failed", "result fail"}
		json.NewEncoder(w).Encode(errmsg)
	}
	if result.Result != "Could not create instance" {
		go cloud.GetNodeState(&result.InstanceID, node, cluster, aKey, sKey)
	}
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
		// fmt.Println("element : ", element)
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
		// fmt.Println("##############", clusterlist)
		// fmt.Println("##############", clusterlist[region])
		clusterlist[region] =
			Region{
				region,
				Attributes{clusterStatus},
				append(clusterlist[region].Children, ChildNode{clustername, Attributes{clusterStatus}})}
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

				if healthCheck["Ready"] == "True" && (healthCheck["NetworkUnavailable"] == "" || (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "False")) && healthCheck["MemoryPressure"] == "False" && healthCheck["DiskPressure"] == "False" && healthCheck["PIDPressure"] == "False" {
					healthyNodeCnt++
				} else {
					if healthCheck["Ready"] == "Unknown" || (healthCheck["NetworkUnavailable"] == "" || healthCheck["NetworkUnavailable"] == "Unknown") || healthCheck["MemoryPressure"] == "Unknown" || healthCheck["DiskPressure"] == "Unknown" || healthCheck["PIDPressure"] == "Unknown" {
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

}

// func WorkloadsDeploymentsOverviewList(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 	w.WriteHeader(http.StatusOK)
// 	if err := json.NewEncoder(w).Encode(resource.ListResource()); err != nil {
// 		panic(err)
// 	}

// }

// func WorkloadsPodsOverviewList(w http.ResponseWriter, r *http.Request) {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	vars := mux.Vars(r)

// 	var client http.Client
// 	resp, err := client.Get("https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/daemonsets/list")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		bodyBytes, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(bodyBytes)
// 	}
// }

// func getDeploymentList(w http.ResponseWriter, r *http.Request) {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	vars := mux.Vars(r)

// 	var client http.Client
// 	resp, err := client.Get("https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/deployments/list")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		bodyBytes, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(bodyBytes)
// 	}
// }

// func getDeploymentDetail(w http.ResponseWriter, r *http.Request) {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	vars := mux.Vars(r)

// 	var client http.Client

// 	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/detail"
// 	//fmt.Print(callUrl)

// 	resp, err := client.Get(callUrl)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		bodyBytes, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(bodyBytes)
// 	}
// }

// func getDeploymentYaml(w http.ResponseWriter, r *http.Request) {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	vars := mux.Vars(r)

// 	var client http.Client

// 	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/yaml"
// 	//fmt.Print(callUrl)

// 	resp, err := client.Get(callUrl)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		bodyBytes, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(bodyBytes)
// 	}
// }

// func getDeploymentEvent(w http.ResponseWriter, r *http.Request) {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	vars := mux.Vars(r)

// 	var client http.Client

// 	callUrl := "https://" + targetURL + "/seedcontainer/api/v1/clusters/" + vars["clusterName"] + "/namespaces/" + vars["namespaceName"] + "/deployments/" + vars["deploymentName"] + "/events"
// 	//fmt.Print(callUrl)

// 	resp, err := client.Get(callUrl)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusOK {
// 		bodyBytes, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(bodyBytes)
// 	}
// }

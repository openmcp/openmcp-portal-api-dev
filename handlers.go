package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"portal-api-server/cloud"
	"portal-api-server/handler"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-11-01/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var openmcpURL = handler.InitPortalConfig()

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

func getAKSVM(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

}

func AKSGetAllResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	clientID := "1edadbd7-d466-43b1-ad73-15a2ee9080ff"
	clientSec := "07.Tx2r7GobBf.Suq7quNRhO_642z-p~6a"
	tenantID := "bc231a1b-ab45-4865-bdba-7724c2893f1c"
	subID := "dc80d3cf-4e1a-4b9a-8785-65c4b739e8d2"

	authBaseURL := azure.PublicCloud.ActiveDirectoryEndpoint
	resourceURL := azure.PublicCloud.ResourceManagerEndpoint
	oauthConfig, err := adal.NewOAuthConfig(authBaseURL, tenantID)

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSec, resourceURL)
	if err != nil {
		fmt.Println("tokenError")
		fmt.Println(err)
		json.NewEncoder(w).Encode(err)
	}

	authorizer := autorest.NewBearerAuthorizer(token)
	ctx := context.Background()

	aksClient := containerservice.NewManagedClustersClientWithBaseURI(resourceURL, subID)
	aksClient.Authorizer = authorizer
	vmssClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(resourceURL, subID)
	vmssClient.Authorizer = authorizer

	var lists []ManagedCluster

	for list, err := aksClient.ListComplete(ctx); list.NotDone(); err = list.Next() {
		if err != nil {
			fmt.Println("got error while traverising Cluster list: ", err)
		}
		clusters := list.Value()

		aPools := *clusters.AgentPoolProfiles
		ap := make(map[string]AgentPool)

		var poolNames []string
		for _, pool := range aPools {
			poolName := *pool.Name
			poolNames = append(poolNames, poolName)
			ap[poolName] = AgentPool{poolName, ""}
		}

		lis := strings.Split(*clusters.ID, "/")
		rgNum := 4
		for index, s := range lis {
			if s == "resourcegroups" {
				rgNum = index + 1
			}
		}
		rg := lis[rgNum]
		nodeRG := *clusters.NodeResourceGroup
		var aplist []AgentPool
		for list, err := vmssClient.ListComplete(ctx, nodeRG); list.NotDone(); err = list.Next() {
			if err != nil {
				fmt.Println("got error while traverising vms list: ", err)
			}
			i := list.Value()
			// fmt.Println(*i.Name)
			poolName := ap[*i.Tags["poolName"]].Name
			vmssName := *i.Name
			aplist = append(aplist, AgentPool{poolName, vmssName})
		}
		lists = append(lists, ManagedCluster{*clusters.Name, rg, nodeRG, aplist, *clusters.Location})
	}
	json.NewEncoder(w).Encode(lists)
}

func AKSNodePower(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	clientID := "1edadbd7-d466-43b1-ad73-15a2ee9080ff"
	clientSec := "07.Tx2r7GobBf.Suq7quNRhO_642z-p~6a"
	tenantID := "bc231a1b-ab45-4865-bdba-7724c2893f1c"
	subID := "dc80d3cf-4e1a-4b9a-8785-65c4b739e8d2"
	// rg := "kube-test-2"
	nodeRG := "MC_kube-test-2_azure-cluster-2_japanwest"
	// targetAgentPoolName := "agentpool2"
	// vmName := "aks-agentpool2-39640967-vmss_0"

	authBaseURL := azure.PublicCloud.ActiveDirectoryEndpoint
	resourceURL := azure.PublicCloud.ResourceManagerEndpoint
	oauthConfig, err := adal.NewOAuthConfig(authBaseURL, tenantID)

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSec, resourceURL)
	if err != nil {
		fmt.Println("tokenError")
		fmt.Println(err)
		json.NewEncoder(w).Encode(err)
	}

	authorizer := autorest.NewBearerAuthorizer(token)

	aksClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(resourceURL, subID)
	aksClient.Authorizer = authorizer
	ctx := context.Background()
	var vmssNames []string

	for list, err := aksClient.ListComplete(ctx, nodeRG); list.NotDone(); err = list.Next() {
		if err != nil {
			fmt.Println("got error while traverising vms list: ", err)
		}
		i := list.Value()

		vmssNames = append(vmssNames, *i.Name)
		fmt.Println(*i.Name)
	}

	vmsClient := compute.NewVirtualMachineScaleSetVMsClientWithBaseURI(resourceURL, subID)
	vmsClient.Authorizer = authorizer

	for list, err := vmsClient.ListComplete(ctx, nodeRG, vmssNames[1], "", "", ""); list.NotDone(); err = list.Next() {
		if err != nil {
			fmt.Println("got error while traverising vms list: ", err)
		}

		i := list.Value()

		fmt.Println(*i.Name, *i.ID)
		progress, err := vmsClient.PowerOff(ctx, nodeRG, vmssNames[1], *i.InstanceID, nil)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(progress)
	}

	// json.NewEncoder(w).Encode(aa)
}

func AKSChangeVMSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	clientID := "1edadbd7-d466-43b1-ad73-15a2ee9080ff"
	clientSec := "07.Tx2r7GobBf.Suq7quNRhO_642z-p~6a"
	tenantID := "bc231a1b-ab45-4865-bdba-7724c2893f1c"
	subID := "dc80d3cf-4e1a-4b9a-8785-65c4b739e8d2"
	// rg := "kube-test-2"
	nodeRG := "MC_kube-test-2_azure-cluster-2_japanwest"
	// targetAgentPoolName := "agentpool2"
	// vmName := "aks-agentpool2-39640967-vmss_0"

	authBaseURL := azure.PublicCloud.ActiveDirectoryEndpoint
	resourceURL := azure.PublicCloud.ResourceManagerEndpoint
	oauthConfig, err := adal.NewOAuthConfig(authBaseURL, tenantID)

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSec, resourceURL)
	if err != nil {
		fmt.Println("tokenError")
		fmt.Println(err)
		json.NewEncoder(w).Encode(err)
	}

	authorizer := autorest.NewBearerAuthorizer(token)

	aksClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(resourceURL, subID)
	aksClient.Authorizer = authorizer
	ctx := context.Background()
	var vmssNames []string
	var locations []string
	var skuCapas []int64
	for list, err := aksClient.ListComplete(ctx, nodeRG); list.NotDone(); err = list.Next() {
		if err != nil {
			fmt.Println("got error while traverising vms list: ", err)
		}
		i := list.Value()

		vmssNames = append(vmssNames, *i.Name)
		locations = append(locations, *i.Location)
		skuCapas = append(skuCapas, *i.Sku.Capacity)
		fmt.Println(*i.Name)
	}
	// // get available Skus
	// for _, vmss := range vmssNames {
	// 	skus, err := aksClient.ListSkus(ctx, nodeRG, vmss)
	// 	if err != nil {
	// 		json.NewEncoder(w).Encode(err)
	// 	}
	// 	json.NewEncoder(w).Encode(skus.Values())
	// }
	skuTierStr := "Standard"
	skuNameStr := "Standard_B1s"
	targetVMSS := vmssNames[0]
	location := locations[0]
	skuCapa := skuCapas[0]
	task, err := aksClient.CreateOrUpdate(
		ctx,
		nodeRG,
		targetVMSS,
		compute.VirtualMachineScaleSet{
			Location: &location,
			Sku: &compute.Sku{
				Tier:     &skuTierStr,
				Name:     &skuNameStr,
				Capacity: &skuCapa,
			},
		},
	)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	json.NewEncoder(w).Encode(task)
}

func AddAKSnode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	clientID := "1edadbd7-d466-43b1-ad73-15a2ee9080ff"
	clientSec := "07.Tx2r7GobBf.Suq7quNRhO_642z-p~6a"
	tenantID := "bc231a1b-ab45-4865-bdba-7724c2893f1c"
	subID := "dc80d3cf-4e1a-4b9a-8785-65c4b739e8d2"
	resourceGroupName := "DefaultResourceGroup-SE"
	resourceName := "azure-cluster-1"
	targetAgentPoolName := "agentpool"

	authBaseURL := azure.PublicCloud.ActiveDirectoryEndpoint
	resourceURL := azure.PublicCloud.ResourceManagerEndpoint
	oauthConfig, err := adal.NewOAuthConfig(authBaseURL, tenantID)

	token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSec, resourceURL)
	if err != nil {
		fmt.Println("tokenError")
		fmt.Println(err)
		json.NewEncoder(w).Encode(err)
	}

	authorizer := autorest.NewBearerAuthorizer(token)
	aksClient := containerservice.NewManagedClustersClientWithBaseURI(resourceURL, subID)
	fmt.Println(resourceURL)

	aksClient.Authorizer = authorizer
	ctx := context.Background()

	// region := "koreacentral"
	// c, err := aksClient.ListOrchestrators(ctx, region, "managedClusters")
	// for _, profile := range *c.Orchestrators {
	// 	fmt.Println(*profile.OrchestratorType)
	// }

	c, err := aksClient.Get(ctx, resourceGroupName, resourceName)
	//	get provision state
	json.NewEncoder(w).Encode(c.ProvisioningState)

	var nodeCount int32
	var location string
	if err != nil {
		fmt.Println("Get AgentPools Error")
		fmt.Println(err)
		json.NewEncoder(w).Encode(err)
	} else {
		location = *c.Location
		for _, profile := range *c.AgentPoolProfiles {
			if *profile.Name == targetAgentPoolName {
				nodeCount = *profile.Count
				break
			}
		}
	}

	nodeCount = nodeCount + 1
	fmt.Println(nodeCount)

	res, err := aksClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceName,
		containerservice.ManagedCluster{
			Location: &location,
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
					{
						Count: to.Int32Ptr(nodeCount),
						Name:  to.StringPtr(targetAgentPoolName),
					},
				},
			},
		},
	)

	json.NewEncoder(w).Encode(res)

	// 	get provision state after change config
	c, err = aksClient.Get(ctx, resourceGroupName, resourceName)
	json.NewEncoder(w).Encode(c.ProvisioningState)
}

func AddEKSnode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	//akid := "AKIAJGFO6OXHRN2H6DSA"                          //
	// secretkey := "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" //
	akid := "AKIAVJTB7UPJPEMHUAJR"
	secretkey := "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" //
	sess, err := session.NewSession(&aws.Config{
		// Region:      aws.String("	ap-northeast-2"), //
		Region:      aws.String("eu-west-2"), //
		Credentials: credentials.NewStaticCredentials(akid, secretkey, ""),
	})

	if err != nil {
		errmsg := jsonErr{503, "failed", "result fail"}
		json.NewEncoder(w).Encode(errmsg)
	}

	svc := eks.New(sess)

	result, err := svc.ListNodegroups(&eks.ListNodegroupsInput{
		ClusterName: aws.String("eks-cluster1"), //
	})

	nodegroupname := result.Nodegroups[0]

	result2, err := svc.DescribeNodegroup(&eks.DescribeNodegroupInput{
		ClusterName:   aws.String("eks-cluster1"), //
		NodegroupName: aws.String(*nodegroupname),
	})

	beforecnt := result2.Nodegroup.ScalingConfig.DesiredSize
	var nodecnt int64
	nodecnt = 1
	desirecnt := *beforecnt + nodecnt

	// // la := make(map[string]*string)
	// // namelabel := "newlabel01"
	// // la["newlabel01"] = &namelabel

	// labelinput := eks.UpdateLabelsPayload{la["newlabel01"]}

	addResult, err := svc.UpdateNodegroupConfig(&eks.UpdateNodegroupConfigInput{
		ClusterName:   aws.String("eks-cluster1"), //
		NodegroupName: aws.String(*nodegroupname),
		// Labels:        &eks.UpdateLabelsPayload{AddOrUpdateLabels: la},
		ScalingConfig: &eks.NodegroupScalingConfig{DesiredSize: &desirecnt},
	})

	if err != nil {
		errmsg := jsonErr{503, "failed", "result fail"}
		json.NewEncoder(w).Encode(errmsg)
	}

	fmt.Println(addResult)
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

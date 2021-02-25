package handler

import (
	"encoding/json"
	"net/http"
)

func GetKVMNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/getkvmnodes?agenturl=192.168.0.96
	agentURL := r.URL.Query().Get("agenturl")

	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/getkvmlists")
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)

}

func StartKVMNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/startkvmnode?agenturl=192.168.0.96&node=rancher
	agentURL := r.URL.Query().Get("agenturl")
	nodeName := r.URL.Query().Get("node")
	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/kvmstartnode?node=" + nodeName)
	if err != nil {

		errorJson := jsonErr{500, "agent connect fail", err.Error()}
		json.NewEncoder(w).Encode(errorJson)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)
}

func StopKVMNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/stopkvmnode?agenturl=192.168.0.96&node=rancher
	agentURL := r.URL.Query().Get("agenturl")
	nodeName := r.URL.Query().Get("node")
	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/kvmstopnode?node=" + nodeName)
	if err != nil {

		errorJson := jsonErr{500, "agent connect fail", err.Error()}
		json.NewEncoder(w).Encode(errorJson)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)
}

func ChangeKVMNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/changekvmnode?agenturl=192.168.0.96&node=rancher&cpu=4&mem=8088
	agentURL := r.URL.Query().Get("agenturl")
	nodeName := r.URL.Query().Get("node")
	vCpu := r.URL.Query().Get("cpu")
	memory := r.URL.Query().Get("mem")
	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/changekvmnode?node=" + nodeName + "&cpu=" + vCpu + "&mem=" + memory)

	if err != nil {
		errorJson := jsonErr{500, "agent connect fail", err.Error()}
		json.NewEncoder(w).Encode(errorJson)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)
}

func CreateKVMNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/createkvmnode?agenturl=192.168.0.96&template=ubuntu16.04-clean&newvm=newvmvmvmvmvmvm
	agentURL := r.URL.Query().Get("agenturl")
	newvm := r.URL.Query().Get("newvm")
	template := r.URL.Query().Get("template")

	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/createkvmnode?template=" + template + "&newvm=" + newvm)

	if err != nil {
		errorJson := jsonErr{500, "agent connect fail", err.Error()}
		json.NewEncoder(w).Encode(errorJson)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)
}

func DeleteKVMNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// http://192.168.0.89:4885/apis/deletekvmnode?agenturl=192.168.0.96&targetvm=newnode-1
	agentURL := r.URL.Query().Get("agenturl")
	targetvm := r.URL.Query().Get("targetvm")

	var client http.Client
	resp, err := client.Get("http://" + agentURL + ":10000/deletekvmnode?node=" + targetvm)

	if err != nil {
		errorJson := jsonErr{500, "agent connect fail", err.Error()}
		json.NewEncoder(w).Encode(errorJson)
	}

	defer resp.Body.Close()

	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	json.NewEncoder(w).Encode(&data)

}

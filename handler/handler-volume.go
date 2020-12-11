package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Persistent Volume Claim (PVC)
// name : items ("metadata","name") string
// status : items ("status","phase") string
// namespace : items ("metatdata","namespace") string
// capacity : items ("status","capacity" storage") string
// createdTime items ("metadata","creationTimestamp")

// persistent volume claim (PVC)
func GetVolumes(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Resultmap)
	token := GetOpenMCPToken()

	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	projectName := vars["projectName"]

	resVolume := VolumeRes{}
	volume := VolumeInfo{}
	volumeURL := "http://" + openmcpURL + "/api/v1/namespaces/" + projectName + "/persistentvolumeclaims?clustername=" + clusterName

	go CallAPI(token, volumeURL, ch)

	volumeResult := <-ch
	volumeData := volumeResult.data
	volumeItems := volumeData["items"].([]interface{})

	if volumeItems != nil {
		for _, element := range volumeItems {
			name := GetStringElement(element, []string{"metadata", "name"})
			namespace := GetStringElement(element, []string{"metadata", "namespace"})
			status := GetStringElement(element, []string{"status", "phase"})
			capacity := GetStringElement(element, []string{"status", "capacity", "storage"})
			createdTime := GetStringElement(element, []string{"metadata", "creationTimestamp"})

			volume.Name = name
			volume.Project = namespace
			volume.Status = status
			volume.Capacity = capacity
			volume.CreatedTime = createdTime

			resVolume.Volumes = append(resVolume.Volumes, volume)
		}
	}
	json.NewEncoder(w).Encode(resVolume.Volumes)
}

package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Resultmap struct {
	secs float64
	url  string
	data map[string]interface{}
}

func GetOpenMCPToken() string {
	var client http.Client

	resp, err := client.Get("http://" + openmcpURL + "/token?username=openmcp&password=keti")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	token := ""

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal([]byte(bodyBytes), &data)
		token = data["token"].(string)

	}
	return token
}

func CallAPI(token string, url string, ch chan<- Resultmap) {
	start := time.Now()
	var bearer = "Bearer " + token
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	var data map[string]interface{}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close() // 리소스 누출 방지
	if err != nil {
		// ch <- fmt.Sprintf("while reading %s: %v", url, err)
		// return
		log.Fatal(err)
	}
	json.Unmarshal([]byte(bodyBytes), &data)

	secs := time.Since(start).Seconds()

	// ch <- fmt.Sprintf("%.2fs %s %v", secs, url, data)

	ch <- Resultmap{secs, url, data}

}

func PercentChange(child, mother float64) (result float64) {
	// diff := float64(new - old)
	result = (float64(child) / float64(mother)) * 100
	return
}

func PercentUseString(child, mother string) (result string) {
	c, _ := strconv.ParseFloat(child, 64)
	m, _ := strconv.ParseFloat(mother, 64)

	if m == 0 || c == 0 {
		return "0.0"
	}
	res := (c / m) * 100
	result = fmt.Sprintf("%.1f", res)
	return
}

func NodeHealthCheck(condType string) string {
	result := ""

	return result
}

func ClusterHealthCheck(condType string) string {
	result := ""

	return result
}

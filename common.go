package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

// func GetDecimalP2 (f []float64) string {

// }

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

func GetJsonBody(rbody io.Reader) map[string]interface{} {
	bodyBytes, err := ioutil.ReadAll(rbody)

	var data map[string]interface{}

	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(bodyBytes), &data)
	return data
}

func PostYaml(url string, yaml io.Reader) ([]byte, error) {
	token := GetOpenMCPToken()
	// fmt.Println("yaml   :", yaml)
	var bearer = "Bearer " + token
	req, err := http.NewRequest("POST", url, yaml)

	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	str := string(respBody)
	fmt.Println(str)
	return respBody, nil

}
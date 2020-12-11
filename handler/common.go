package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/jinzhu/configor"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var InfluxConfig = struct {
	Influx struct {
		Ip       string
		Port     string
		Username string
		Password string
	}
}{}

func InitInfluxConfig() {
	configor.Load(&InfluxConfig, "dbconfig.yml")
}

type Influx struct {
	inClient client.Client
}

func NewInflux(INFLUX_IP, INFLUX_PORT, username, password string) *Influx {
	inf := &Influx{
		inClient: InfluxDBClient(INFLUX_IP, INFLUX_PORT, username, password),
	}
	return inf
}

func InfluxDBClient(INFLUX_IP, INFLUX_PORT, username, password string) client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://" + INFLUX_IP + ":" + INFLUX_PORT,
		Username: username,
		Password: password,
	})
	if err != nil {
		fmt.Println(err)
	}
	return c
}

type jsonErr struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
	Text   string `json:"text"`
}

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
		// log.Fatal(err)
		fmt.Print(err)
	}
	var data map[string]interface{}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close() // 리소스 누출 방지
	if err != nil {
		// ch <- fmt.Sprintf("while reading %s: %v", url, err)
		// return
		fmt.Print(err)
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

func GetInfluxPodsMetric(clusterName string, in *Influx) []client.Result {
	q := client.Query{}
	// q = client.NewQuery("SELECT last(*) FROM Pods WHERE cluster = '"+clusterName+"' ORDER BY DESC LIMIT 1", "Metrics", "")
	// select last(*) from Pods where time > now() - 5m and cluster='cluster1' group by namespace,pod order by desc limit 1
	q = client.NewQuery("select last(*) from Pods where time > now() - 5m and cluster='"+clusterName+"' group by namespace,pod order by desc limit 1", "Metrics", "")

	//select last(*) from Pods where time > now() -1m and cluster='cluster1' group by namespace,pod,time(1m)
	response, err := in.inClient.Query(q)

	if err == nil && response.Error() == nil {

		return response.Results
	}

	return nil
}

// func GetInfluxPod10mMetric(clusterName string, namespace string, pod string, in *Influx) []client.Result {
func GetInfluxPod10mMetric(clusterName string, namespace string, pod string) PhysicalResources {
	nowTime := time.Now().UTC() //.Add(time.Duration(offset) * time.Second)
	startTime := nowTime.Add(time.Duration(-10) * time.Minute)
	endTime := nowTime
	_, offset := time.Now().Zone()
	start := startTime.Format("2006-01-02_15:04:05")
	end := endTime.Format("2006-01-02_15:04:05")

	ch := make(chan Resultmap)
	token := GetOpenMCPToken()
	// http://192.168.0.152:31635/metrics/namespaces/kube-system/pods/kube-flannel-ds-nn5p5?clustername=cluster1&timeStart=2020-09-03_09:00:00&timeEnd=2020-09-03_09:00:15

	podMetricURL := "http://" + openmcpURL + "/metrics/namespaces/" + namespace + "/pods/" + pod + "?clustername=" + clusterName + "&timeStart=" + start + "&timeEnd=" + end
	go CallAPI(token, podMetricURL, ch)

	podMetricResult := <-ch
	podMetricData := podMetricResult.data["podmetrics"]

	metricsPerMin := make(map[string][]interface{})
	for _, m := range podMetricData.([]interface{}) {
		times := m.(map[string]interface{})["time"].(string)
		ind := strings.Index(times, ":")
		timeHM := times[ind-2 : ind+3]
		timeHM = timeHM + ":00"
		t1, _ := time.Parse("15:04:05", timeHM)
		t1 = t1.Add(time.Duration(offset) * time.Second)
		timeHM = t1.Format("15:04:05")

		metricsPerMin[timeHM] = append(metricsPerMin[timeHM], m)
	}

	var podCPUUsageMins []PodCPUUsageMin
	var podMemoryUsageMins []PodMemoryUsageMin
	var podNetworkUsageMins []PodNetworkUsageMin
	for k, m := range metricsPerMin {
		cpuSum := 0
		memorySum := 0
		oldNtTxUseInt := 0
		oldNtRxUseInt := 0
		maxTxUseInt := 0
		minTxUseInt := 0
		maxRxUseInt := 0
		minRxUseInt := 0

		for index, v := range m {
			if v.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"] != nil {
				cpuUse := v.(map[string]interface{})["cpu"].(map[string]interface{})["CPUUsageNanoCores"].(string)
				cpuUse = strings.Split(cpuUse, "n")[0]
				cpuUseInt, _ := strconv.Atoi(cpuUse)
				cpuSum += cpuUseInt
			}

			if v.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"] != nil {
				memoryUse := v.(map[string]interface{})["memory"].(map[string]interface{})["MemoryUsageBytes"].(string)
				memoryUse = strings.Split(memoryUse, "Ki")[0]
				memoryUseInt, _ := strconv.Atoi(memoryUse)
				memorySum += memoryUseInt
			}
			ntTxUseInt := 0
			ntRxUseInt := 0
			if v.(map[string]interface{})["network"].(map[string]interface{})["NetworkTxBytes"] != nil {
				ntTxUse := v.(map[string]interface{})["network"].(map[string]interface{})["NetworkTxBytes"].(string)
				ntTxUseInt, _ = strconv.Atoi(ntTxUse)
			}

			if v.(map[string]interface{})["network"].(map[string]interface{})["NetworkRxBytes"] != nil {
				ntRxUse := v.(map[string]interface{})["network"].(map[string]interface{})["NetworkRxBytes"].(string)
				ntRxUseInt, _ = strconv.Atoi(ntRxUse)
			}
			// fmt.Println(v.(map[string]interface{})["time"], ntTxUseInt, ntRxUseInt)
			if index == 0 {
				oldNtTxUseInt = ntTxUseInt
				oldNtRxUseInt = ntRxUseInt
				minTxUseInt = ntTxUseInt
				minRxUseInt = ntRxUseInt
			} else {
				if oldNtTxUseInt < ntTxUseInt {
					maxTxUseInt = ntTxUseInt
				}
				if oldNtRxUseInt < ntRxUseInt {
					maxRxUseInt = ntRxUseInt
				}

				oldNtTxUseInt = ntTxUseInt
				oldNtRxUseInt = ntRxUseInt
			}
		}

		cpuAvg := float64(cpuSum) / float64(len(m)) / 1000 / 1000 / 1000
		memoryAvg := float64(memorySum) / float64(len(m)) / 1000
		inBps := (maxTxUseInt - minTxUseInt) / 60
		outBps := (maxRxUseInt - minRxUseInt) / 60
		//fmt.Println(k, "cpu: ", cpuAvg)
		podCPUUsageMins = append(podCPUUsageMins, PodCPUUsageMin{math.Ceil(cpuAvg*1000) / 1000, k})
		podMemoryUsageMins = append(podMemoryUsageMins, PodMemoryUsageMin{math.Ceil(memoryAvg*10) / 10, k})
		podNetworkUsageMins = append(podNetworkUsageMins, PodNetworkUsageMin{"Bps", inBps, outBps, k})

	}
	sort.Slice(podCPUUsageMins, func(i, j int) bool {
		return podCPUUsageMins[i].Time < podCPUUsageMins[j].Time
	})
	sort.Slice(podMemoryUsageMins, func(i, j int) bool {
		return podMemoryUsageMins[i].Time < podMemoryUsageMins[j].Time
	})
	sort.Slice(podNetworkUsageMins, func(i, j int) bool {
		return podNetworkUsageMins[i].Time < podNetworkUsageMins[j].Time
	})

	if len(podCPUUsageMins) > 10 {
		podCPUUsageMins = podCPUUsageMins[1:]
		podMemoryUsageMins = podMemoryUsageMins[1:]
		podNetworkUsageMins = podNetworkUsageMins[1:]
	}
	result := PhysicalResources{podCPUUsageMins, podMemoryUsageMins, podNetworkUsageMins}
	return result
}

func reverseRank(data map[string]float64, top int) PairList {
	pl := make(PairList, len(data))

	if top > len(data) {
		top = len(data)
	}
	i := 0
	for k, v := range data {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl[:top]
}

type Pair struct {
	Name  string  `json:"name"`
	Usage float64 `json:"usage"`
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Usage < p[j].Usage }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func Round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func GetStringElement(nMap interface{}, keys []string) string {
	result := ""
	if nMap.(map[string]interface{})[keys[0]] != nil {
		childMap := nMap.(map[string]interface{})[keys[0]]
		for i, _ := range keys {
			typeCheck := fmt.Sprintf("%T", childMap)

			if len(keys)-1 == i {
				if "[]interface {}" == typeCheck {
					result = childMap.([]interface{})[0].(string)
				} else {
					result = childMap.(string)
				}
				break
			}

			if "[]interface {}" == typeCheck {
				if childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]]
				} else {
					result = "-"
					break
				}
			} else {
				if childMap.(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.(map[string]interface{})[keys[i+1]]
				} else {
					result = "-"
					break
				}
			}
		}
	} else {
		result = "-"
	}
	return result
}

func GetIntElement(nMap interface{}, keys []string) int {
	result := 0
	if nMap.(map[string]interface{})[keys[0]] != nil {
		childMap := nMap.(map[string]interface{})[keys[0]]
		for i, _ := range keys {
			typeCheck := fmt.Sprintf("%T", childMap)

			if len(keys)-1 == i {
				if "[]interface {}" == typeCheck {
					result = childMap.([]interface{})[0].(int)
				} else {
					result = childMap.(int)
				}
				break
			}

			if "[]interface {}" == typeCheck {
				if childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]]
				} else {
					result = 0
					break
				}
			} else {
				if childMap.(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.(map[string]interface{})[keys[i+1]]
				} else {
					result = 0
					break
				}
			}
		}
	} else {
		result = 0
	}
	return result
}

func GetFloat64Element(nMap interface{}, keys []string) float64 {
	var result float64 = 0.0
	if nMap.(map[string]interface{})[keys[0]] != nil {
		childMap := nMap.(map[string]interface{})[keys[0]]
		for i, _ := range keys {
			typeCheck := fmt.Sprintf("%T", childMap)

			if len(keys)-1 == i {
				if "[]interface {}" == typeCheck {
					result = childMap.([]interface{})[0].(float64)
				} else {
					result = childMap.(float64)
				}
				break
			}

			if "[]interface {}" == typeCheck {
				if childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]]
				} else {
					result = 0.0
					break
				}
			} else {
				if childMap.(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.(map[string]interface{})[keys[i+1]]
				} else {
					result = 0.0
					break
				}
			}
		}
	} else {
		result = 0.0
	}
	return result
}

func GetInterfaceElement(nMap interface{}, keys []string) interface{} {
	var result interface{}
	if nMap.(map[string]interface{})[keys[0]] != nil {
		childMap := nMap.(map[string]interface{})[keys[0]]
		for i, _ := range keys {
			typeCheck := fmt.Sprintf("%T", childMap)

			if len(keys)-1 == i {
				if "[]interface {}" == typeCheck {
					result = childMap.([]interface{})[0]
				} else {
					result = childMap
				}
				break
			}

			if "[]interface {}" == typeCheck {
				if childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]]
				} else {
					result = nil
					break
				}
			} else {
				if childMap.(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.(map[string]interface{})[keys[i+1]]
				} else {
					result = nil
					break
				}
			}
		}
	} else {
		result = nil
	}
	return result
}

func GetArrayElement(nMap interface{}, keys []string) []interface{} {
	var result []interface{}
	if nMap.(map[string]interface{})[keys[0]] != nil {
		childMap := nMap.(map[string]interface{})[keys[0]]
		for i, _ := range keys {
			typeCheck := fmt.Sprintf("%T", childMap)

			if len(keys)-1 == i {
				if "[]interface {}" == typeCheck {
					result = childMap.([]interface{})
				} else {
					result = childMap.([]interface{})
				}
				break
			}

			if "[]interface {}" == typeCheck {
				if childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.([]interface{})[0].(map[string]interface{})[keys[i+1]]
				} else {
					result = nil
					break
				}
			} else {
				if childMap.(map[string]interface{})[keys[i+1]] != nil {
					childMap = childMap.(map[string]interface{})[keys[i+1]]
				} else {
					result = nil
					break
				}
			}
		}
	} else {
		result = nil
	}
	return result
}

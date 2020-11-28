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

func GetInfluxPodsMetric(clusterName string, in *Influx) []client.Result {
	q := client.Query{}
	// q = client.NewQuery("SELECT last(*) FROM Pods WHERE cluster = '"+clusterName+"' ORDER BY DESC LIMIT 1", "Metrics", "")
	q = client.NewQuery("select last(*) from Pods where time > now() - 1m and cluster='"+clusterName+"' group by namespace,pod,time(1m) order by desc limit 1", "Metrics", "")
	//select last(*) from Pods where time > now() -1m and cluster='cluster1' group by namespace,pod,time(1m)
	response, err := in.inClient.Query(q)

	if err == nil && response.Error() == nil {

		return response.Results
	}

	return nil
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

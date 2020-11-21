package executors

import (
	"fmt"
	client "github.com/bozd4g/go-http-client"
	"net/http"
	"pigeon/db"
	"time"
)

func GetRequest(taskId string, continuous bool, url string, header map[string]interface{}) {
	httpClient := client.New(url)
	req, err := httpClient.Get("")

	for i, h := range header {
		hv := fmt.Sprintf("%v", h)
		req.Header.Add(i, hv)
	}

	if err != nil {
		println(err)
	}

	doRequest(taskId, continuous, httpClient, req)
}

func PostRequest(taskId string, continuous bool, url string, body map[string]interface{}, header map[string]interface{}) {
	httpClient := client.New(url)
	req, err := httpClient.PostWith("", body)

	for i, h := range header {
		hv := fmt.Sprintf("%v", h)
		req.Header.Add(i, hv)
	}

	if err != nil {
		println(err)
	}

	doRequest(taskId, continuous, httpClient, req)
}

func doRequest(taskId string, continuous bool, httpClient client.IHttpClient, req *http.Request) {
	if db.CheckExecutionAvailability(taskId) {
		resp := httpClient.Do(req)
		if resp.IsSuccess {
			fmt.Println("success for " + taskId + " - send at : " + time.Now().String())
			if !continuous {
				db.DecreaseReamingCount(taskId)
			}
		} else {
			fmt.Println(resp.Error)
		}
	}
}

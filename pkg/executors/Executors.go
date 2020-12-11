package executors

import (
	"context"
	"encoding/json"
	"fmt"
	client "github.com/bozd4g/go-http-client"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"pigeon/db"
	pb "pigeon/proto"
	"time"
)

func GetRequest(taskId string, url string, header map[string]interface{}) {
	log.Printf("[GET] executed for %v", taskId)

	httpClient := client.New(url)
	req, err := httpClient.Get("")

	for i, h := range header {
		hv := fmt.Sprintf("%v", h)
		req.Header.Add(i, hv)
	}

	if err != nil {
		println(err)
	}

	doRequest(taskId, httpClient, req)
}

func PostRequest(taskId string, url string, body map[string]interface{}, header map[string]interface{}) {
	log.Printf("[POST] executed for %v", taskId)

	httpClient := client.New(url)
	req, err := httpClient.PostWith("", body)

	for i, h := range header {
		hv := fmt.Sprintf("%v", h)
		req.Header.Add(i, hv)
	}

	if err != nil {
		println(err)
	}

	doRequest(taskId, httpClient, req)
}

func PatchRequest(taskId string, url string, body map[string]interface{}, header map[string]interface{}) {
	log.Printf("[PATCH] executed for %v", taskId)

	httpClient := client.New(url)
	req, err := httpClient.PatchWith("", body)

	for i, h := range header {
		hv := fmt.Sprintf("%v", h)
		req.Header.Add(i, hv)
	}

	if err != nil {
		println(err)
	}

	doRequest(taskId, httpClient, req)
}

func doRequest(taskId string, httpClient client.IHttpClient, req *http.Request) {
	if db.CheckExecutionAvailability(taskId) {
		resp := httpClient.Do(req)
		if resp.IsSuccess {
			fmt.Println("success for " + taskId + " - send at : " + time.Now().String() + " ----> " + req.URL.String())
			db.IncreaseFireCount(taskId)
		} else {
			fmt.Println(resp.Error)
			db.UpdateLastFire(taskId)
		}
	}
}

func TriggerScheduledNotification(taskId string, executionBody map[string]interface{}) {
	body, _ := json.Marshal(executionBody)
	req := pb.TriggerNotificationRequest{}
	json.Unmarshal(body, &req)

	conn, err := grpc.Dial(db.GetTargetGrpcServer(), grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	client := pb.NewNotificationServiceClient(conn)
	request := &pb.TriggerNotificationRequest{NotificationId: req.NotificationId}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.TriggerNotification(ctx, request)
	if err != nil {
		db.UpdateLastFire(taskId)
		fmt.Println(err)
	}

	db.IncreaseFireCount(taskId)
	fmt.Println("Response:", response.Done)
}

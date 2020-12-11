package handlers

import (
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"pigeon/db"
	. "pigeon/model"
	. "pigeon/pkg/executors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var s = gocron.NewScheduler(time.Local)

func AttachNewTask(c echo.Context) (err error) {
	req := new(SchedulerRequest)
	if err := c.Bind(&req); err != nil {
		panic(err)
	}

	response, err := attach(*req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

//todo add request validator
func attach(req SchedulerRequest) (response TaskCreatedResponse, err error) {
	fmt.Println("attach")
	taskId := guuid.New().String()

	_, err = prepareTask("NEW", taskId, req.Interval, req.IntervalType, req.SendAt, req.Immediately,
		req.Continuous, req.Limit, 0, time.Now(), req.Execution)
	if err != nil {
		log.Println(err)
	}

	db.Save(taskId, req)

	response = TaskCreatedResponse{
		TaskId: taskId,
	}
	return response, err
}

func GetTaskDetail(c echo.Context) (err error) {
	task := db.GetById(c.Param("id"))
	return c.JSON(http.StatusOK, task)
}

func GetTasks(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, db.GetActives())
}

func CancelTask(c echo.Context) (err error) {
	taskId := c.Param("id")
	db.CancelTask(taskId)
	return s.RemoveJobByTag(taskId)
}

func limitToTask(tags []string, limit int) {
	if s.Len() > 0 {
		jobs := s.Jobs()
		for _, job := range jobs {
			if reflect.DeepEqual(job.Tags(), tags) {
				job.LimitRunsTo(limit)
			}
		}
	}
}

func prepareTask(taskType string, taskId string, interval int, intervalType string, sendAt string, immediately bool,
	continuous bool, limit int, fireCount int, startAt time.Time, exec ExecutionRequest) (job *gocron.Job, err error) {
	_interval := uint64(interval)

	if strings.EqualFold(exec.TargetUrl, "") {
		return nil, errors.New("Target URL does not be empty")
	}

	if !strings.Contains(exec.TargetUrl, "http") {
		return nil, errors.New("Target URL is not acceptable URL")
	}

	s.StartAsync()
	s.Every(_interval)

	// stores public uuid in the job tag
	tags := []string{taskId}
	s.SetTag(tags)

	// starts job when scheduler app restarted, so we need check this situation
	if taskType == "NEW" && immediately {
		s.StartImmediately()
	}

	//todo timezone
	if sendAt == "" {
		hour := strconv.Itoa(time.Now().Hour())

		minute := time.Now().Minute()
		if minute < 10 {
			sendAt = hour + ":" + "0" + strconv.Itoa(minute)
		} else {
			sendAt = hour + ":" + strconv.Itoa(time.Now().Minute())
		}
	}

	switch intervalType {
	case "SECOND":
		s.Second().At(sendAt)
		break
	case "DAILY":
		s.Day().At(sendAt)
		break
	case "WEEKLY":
		//TODO WIP
		break
	default:
		break
	}

	if !continuous {
		if taskType == "LEGACY" {
			limitToTask(tags, limit-fireCount)
		} else {
			limitToTask(tags, limit)
		}
	}

	switch exec.Type {
	case "GET":
		do, err := s.Do(GetRequest, taskId, exec.TargetUrl, exec.Header)
		if err != nil {
			return nil, errors.New("An error occurred while request [GET]")
		}
		return do, err
	case "POST":
		do, err := s.Do(PostRequest, taskId, exec.TargetUrl, exec.Body, exec.Header)
		if err != nil {
			return nil, errors.New("An error occurred while request [POST]")
		}
		return do, err
	case "PATCH":
		do, err := s.Do(PatchRequest, taskId, exec.TargetUrl, exec.Body, exec.Header)
		if err != nil {
			return nil, errors.New("An error occurred while request [PATCH]")
		}
		return do, err
	case "GRPC":
		do, err := s.Do(TriggerScheduledNotification, taskId, exec.Body)
		if err != nil {
			return nil, errors.New("An error occurred while request [GRPC]")
		}
		return do, err
	}

	do, err := s.Do(unrecognizedExecutionType)
	if err != nil {

		return nil, errors.New("Task execution type can not recognize")
	}
	return do, err
}

func QneTimeScheduledNotification(notificationId string, sendAt string) {
	preparedRequest := SchedulerRequest{
		Continuous:   false,
		Immediately:  false,
		Interval:     1,
		IntervalType: "DAILY",
		SendAt:       sendAt,
		Limit:        1,
		Execution: ExecutionRequest{
			TargetUrl: "http://",
			Type:      "GRPC",
			Body: map[string]interface{}{
				"notificationId": notificationId,
			},
			Header: map[string]interface{}{},
		},
	}

	attach(preparedRequest)
}

func unrecognizedExecutionType() {
	fmt.Println("Task execution type can not recognize")
}

func RestartExistsJobs(jobs []TaskDetail) {
	for _, job := range jobs {
		if job.Continuous || (job.Limit-job.SuccessfulFireCount > 0) {
			log.Printf("[%v] will execute", job.TaskId.String())
			prepareTask("LEGACY", job.TaskId.String(), job.Interval, job.IntervalType, job.SendAt, job.Immediately,
				job.Continuous, job.Limit, job.FireCount, job.StartAt, job.Execution)
		}
	}
}

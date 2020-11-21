package handlers

import (
	"fmt"
	"github.com/go-co-op/gocron"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo"
	"net/http"
	"pigeon/db"
	. "pigeon/model"
	. "pigeon/pkg/executors"
	"reflect"
	"time"
)

var s = gocron.NewScheduler(time.Local)

func AttachNewTask(c echo.Context) (err error) {
	req := new(SchedulerRequest)
	if err := c.Bind(&req); err != nil {
		panic(err)
	}

	response, err := attach(*req)
	return c.JSON(http.StatusOK, response)
}

func attach(req SchedulerRequest) (response TaskCreatedResponse, err error) {
	taskId := guuid.New().String()

	_, err = prepareTask("NEW", taskId, req.Interval, req.IntervalType, req.SendAt, req.Immediately, req.Continuous, req.Limit, req.Limit, req.StartAt.ConvertToTime(), req.Execution)
	if err != nil {
		println(err)
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
	jobs := s.Jobs()
	for _, job := range jobs {
		if reflect.DeepEqual(job.Tags(), tags) {
			job.LimitRunsTo(limit)
		}
	}
}

func prepareTask(taskType string, taskId string, interval int, intervalType string, sendAt string, immediately bool,
	continuous bool, limit int, remainCount int, startAt time.Time, exec ExecutionRequest) (job *gocron.Job, err error) {
	_interval := uint64(interval)

	s.StartAsync()
	s.Every(_interval)

	switch intervalType {
	case "SECOND":
		s.Second()
		break
	case "DAILY":
		s.Day()
		break
	case "WEEKLY":
		s.Week()
		break
	default:
		break
	}

	tags := []string{taskId}
	s.SetTag(tags).StartAt(startAt).At(sendAt)

	if immediately {
		s.StartImmediately()
	}

	if !continuous {
		if taskType == "LEGACY" {
			limitToTask(tags, remainCount)
		} else {
			limitToTask(tags, limit)
		}
	}

	switch exec.Type {
	case "GET":
		return s.Do(GetRequest, taskId, continuous, exec.TargetUrl, exec.Header)
		break
	case "POST":
		return s.Do(PostRequest, taskId, continuous, exec.TargetUrl, exec.Body, exec.Header)
		break
	}

	return s.Do(unrecognizedExecutionType)
}

func unrecognizedExecutionType() {
	fmt.Println("Task execution type can not recognize")
}

func RestartExistsJobs(jobs []TaskDetail) {
	for _, job := range jobs {
		prepareTask("LEGACY", job.TaskId.String(), job.Interval, job.IntervalType, job.SendAt, job.Immediately, job.Continuous, job.Limit, job.RemainCount, job.StartAt, job.Execution)
	}
}

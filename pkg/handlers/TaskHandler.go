package handlers

import (
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo"
	"net/http"
	"pigeon/db"
	. "pigeon/model"
	. "pigeon/pkg/executors"
	"reflect"
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

func attach(req SchedulerRequest) (response TaskCreatedResponse, err error) {
	taskId := guuid.New().String()

	_, err = prepareTask("NEW", taskId, req.Interval, req.IntervalType, req.SendAt, req.Immediately,
		req.Continuous, req.Limit, 0, req.StartAt.ConvertToTime(), req.Execution)

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
		return nil,  errors.New("Target URL does not be empty")

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
			return nil, errors.New("An error occurred while request")
		}
		return do, err
	case "POST":
		do, err := s.Do(PostRequest, taskId, exec.TargetUrl, exec.Body, exec.Header)
		if err != nil {
			return nil, errors.New("An error occurred while request")
		}
		return do, err
	}

	do, err := s.Do(unrecognizedExecutionType)
	if err != nil {

		return nil, errors.New("Task execution type can not recognize")
	}
	return do, err
}

func unrecognizedExecutionType() {
	fmt.Println("Task execution type can not recognize")
}

func RestartExistsJobs(jobs []TaskDetail) {
	for _, job := range jobs {
		if job.Continuous || (job.Limit-job.FireCount > 0) {
			prepareTask("LEGACY", job.TaskId.String(), job.Interval, job.IntervalType, job.SendAt, job.Immediately,
				job.Continuous, job.Limit, job.FireCount, job.StartAt, job.Execution)
		}
	}
}

package model

import (
	guuid "github.com/google/uuid"
	"time"
)

type SchedulerStart struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

type ExecutionRequest struct {
	TargetUrl string                 `json:"targetUrl"`
	Type      string                 `json:"type"`
	Body      map[string]interface{} `json:"body"`
	Header    map[string]interface{} `json:"header"`
}

type SchedulerRequest struct {
	Interval     int              `json:"interval"`
	IntervalType string           `json:"intervalType"`
	SendAt       string           `json:"sendAt"`
	StartAt      SchedulerStart   `json:"startAt"`
	Limit        int              `json:"limit"`
	Immediately  bool             `json:"immediately"`
	Continuous   bool             `json:"continuous"`
	Execution    ExecutionRequest `json:"execution"`
}

type TaskDetail struct {
	Id           int64
	TaskId       guuid.UUID
	Interval     int
	IntervalType string
	SendAt       string
	StartAt      time.Time
	Execution    ExecutionRequest
	Limit        int
	Immediately  bool
	Continuous   bool
	Cancelled    bool
	RemainCount  int
}

type TaskSummary struct {
	TaskId        guuid.UUID
	TargetUrl     string
	ExecutionType string
}

type TaskCreatedResponse struct {
	TaskId string
}

func (date SchedulerStart) ConvertToTime() time.Time {
	return time.Date(date.Year, time.Month(date.Month), date.Day, date.Hour, date.Minute, date.Second, 0, time.Local)
}

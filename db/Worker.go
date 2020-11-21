package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	. "pigeon/model"
)

func Save(taskId string, req SchedulerRequest) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// convert map field to JSON field
	eb, err := json.Marshal(req.Execution.Body)
	ebs := string(eb)

	h, err := json.Marshal(req.Execution.Header)
	hs := string(h)

	saveQuery := "INSERT INTO tasks(uuid, start_at, target_url, interval_time, interval_type, send_at, exec_type, exec_body, exec_header, exec_limit, immediately, continuous, cancelled, remain_count) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)"
	_, err = db.Exec(saveQuery, taskId, req.StartAt.ConvertToTime(), req.Execution.TargetUrl, req.Interval, req.IntervalType, req.SendAt, req.Execution.Type, ebs, hs, req.Limit, req.Immediately, req.Continuous, false, req.Limit)
	if err != nil {
		print(err)
	}
}

func GetById(taskId string) (response TaskDetail) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := "SELECT * FROM tasks WHERE uuid = $1"

	var detail TaskDetail
	var execBody string
	var execHeader string

	err = db.QueryRow(findQuery, taskId).Scan(&detail.Id, &detail.TaskId, &detail.StartAt, &detail.Execution.TargetUrl,
		&detail.Interval, &detail.IntervalType, &detail.SendAt, &detail.Execution.Type, &execBody, &execHeader,
		&detail.Limit, &detail.Immediately, &detail.Continuous, &detail.Cancelled, &detail.RemainCount)

	// convert marshalled JSON field
	var eb map[string]interface{}
	if err = json.Unmarshal([]byte(execBody), &eb); err != nil {
		print(err)
	}
	detail.Execution.Body = eb

	var h map[string]interface{}
	if err = json.Unmarshal([]byte(execHeader), &h); err != nil {
		print(err)
	}
	detail.Execution.Header = h

	if err != nil {
		print(err)
	}
	return detail
}

func CancelTask(taskId string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := "UPDATE tasks SET cancelled = true WHERE uuid = $1"

	_, err = db.Exec(query, taskId)
	if err != nil {
		println("an error occurred while cancelled update operation")
	}
}

func GetActives() (response []TaskSummary) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := "SELECT uuid, exec_type, target_url FROM tasks WHERE cancelled = false"

	rows, err := db.Query(findQuery)
	defer rows.Close()

	for rows.Next() {
		summary := TaskSummary{}
		err = rows.Scan(&summary.TaskId, &summary.ExecutionType, &summary.TargetUrl)
		if err != nil {
			panic(err)
		}
		fmt.Println(summary)

		response = append(response, summary)
	}

	return response
}

func GetDetailedJobs() (response []TaskDetail) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := "SELECT * FROM tasks WHERE cancelled = false OR remain_count > 0"

	rows, err := db.Query(findQuery)
	defer rows.Close()

	for rows.Next() {
		detail := TaskDetail{}
		var execBody string
		var execHeader string

		err = rows.Scan(&detail.Id, &detail.TaskId, &detail.StartAt, &detail.Execution.TargetUrl, &detail.Interval,
			&detail.IntervalType, &detail.SendAt, &detail.Execution.Type, &execBody, &execHeader, &detail.Limit,
			&detail.Immediately, detail.Continuous, &detail.Cancelled, &detail.RemainCount)

		// convert marshalled JSON field
		var eb map[string]interface{}
		if err = json.Unmarshal([]byte(execBody), &eb); err != nil {
			print(err)
		}
		detail.Execution.Body = eb

		var h map[string]interface{}
		if err = json.Unmarshal([]byte(execHeader), &h); err != nil {
			print(err)
		}
		detail.Execution.Header = h

		response = append(response, detail)
	}
	return response
}

func DecreaseReamingCount(taskId string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := "UPDATE tasks SET remain_count = remain_count - 1 WHERE uuid = $1"
	_, err = db.Exec(query, taskId)
	if err != nil {
		println("an error occurred while remain_count update operation")
	}
}

func CheckExecutionAvailability(taskId string) bool {
	task := GetById(taskId)

	if task.Limit == 0 {
		return true
	}

	return task.RemainCount > 0
}

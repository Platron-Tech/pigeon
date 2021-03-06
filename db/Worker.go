package db

import (
	"database/sql"
	"encoding/json"
	. "pigeon/model"
	"strconv"
	"time"
)

func Save(taskId string, req SchedulerRequest) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// convert map field to JSON field
	eb, err := json.Marshal(req.Execution.Body)
	ebs := string(eb)

	h, err := json.Marshal(req.Execution.Header)
	hs := string(h)

	if req.SendAt == "" {
		hour := strconv.Itoa(time.Now().Hour())

		minute := time.Now().Minute()
		if minute < 10 {
			req.SendAt = hour + ":" + "0" + strconv.Itoa(minute)
		} else {
			req.SendAt = hour + ":" + strconv.Itoa(time.Now().Minute())
		}
	}

	startAt := time.Now()
	//if req.StartAt != (SchedulerStart{}) { //todo fix later
	//	startAt = req.StartAt.ConvertToTime()
	//}

	saveQuery :=
		`INSERT INTO tasks(uuid, 
							start_at, 
							target_url, 
							interval_time, 
							interval_type, 
							send_at, 
							exec_type, 
							exec_body, 
							exec_header, 
							exec_limit, 
							immediately, 
							continuous, 
							cancelled, 
							fire_count, 
							successful_fire_count, 
							last_fire, 
							last_successful_fire) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`
	_, err = db.Exec(saveQuery, taskId, startAt, req.Execution.TargetUrl, req.Interval, req.IntervalType, req.SendAt,
		req.Execution.Type, ebs, hs, req.Limit, req.Immediately, req.Continuous, false, 0, 0, nil, nil)
	if err != nil {
		print(err)
	}
}

func GetById(taskId string) (response TaskDetail) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := "SELECT * FROM tasks WHERE uuid = $1"

	var detail TaskDetail
	var execBody string
	var execHeader string

	err = db.QueryRow(findQuery, taskId).Scan(&detail.Id,
		&detail.TaskId,
		&detail.StartAt,
		&detail.Execution.TargetUrl,
		&detail.Interval,
		&detail.IntervalType,
		&detail.SendAt,
		&detail.Execution.Type,
		&execBody,
		&execHeader,
		&detail.Limit,
		&detail.Immediately,
		&detail.Continuous,
		&detail.Cancelled,
		&detail.FireCount,
		&detail.SuccessfulFireCount,
		&detail.LastFire,
		&detail.LastSuccessfulFire)

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
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := `UPDATE tasks SET cancelled = true WHERE uuid = $1`

	_, err = db.Exec(query, taskId)
	if err != nil {
		println("an error occurred while cancelled update operation")
	}
}

func GetActives() (response []TaskSummary) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := `SELECT uuid, exec_type, target_url FROM tasks WHERE cancelled = false`

	rows, err := db.Query(findQuery)
	defer rows.Close()

	for rows.Next() {
		summary := TaskSummary{}
		err = rows.Scan(&summary.TaskId, &summary.ExecutionType, &summary.TargetUrl)
		if err != nil {
			panic(err)
		}

		response = append(response, summary)
	}

	return response
}

func GetDetailedJobs() (response []TaskDetail) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	findQuery := `SELECT * FROM tasks WHERE cancelled = false`

	rows, err := db.Query(findQuery)
	defer rows.Close()

	for rows.Next() {
		detail := TaskDetail{}
		var execBody string
		var execHeader string

		err = rows.Scan(&detail.Id,
			&detail.TaskId,
			&detail.StartAt,
			&detail.Execution.TargetUrl,
			&detail.Interval,
			&detail.IntervalType,
			&detail.SendAt,
			&detail.Execution.Type,
			&execBody,
			&execHeader,
			&detail.Limit,
			&detail.Immediately,
			&detail.Continuous,
			&detail.Cancelled,
			&detail.FireCount,
			&detail.SuccessfulFireCount,
			&detail.LastFire,
			&detail.LastSuccessfulFire)

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

func IncreaseFireCount(taskId string) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	now := time.Now()
	query := "UPDATE tasks SET successful_fire_count = successful_fire_count + 1, last_fire = $1, last_successful_fire = $2 WHERE uuid = $3"
	_, err = db.Exec(query, now, now, taskId)
	if err != nil {
		println("an error occurred while successful_fire_count update operation")
	}
}

func UpdateLastFire(taskId string) {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	now := time.Now()
	query := "UPDATE tasks SET fire_count = fire_count + 1, last_fire = $1 WHERE uuid = $2"
	_, err = db.Exec(query, now, taskId)
	if err != nil {
		println("an error occurred while fire_count update operation")
	}
}

func CheckExecutionAvailability(taskId string) bool {
	task := GetById(taskId)

	if task.Continuous {
		return true
	}

	return task.Limit-task.SuccessfulFireCount > 0
}

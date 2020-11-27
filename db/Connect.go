package db

import (
	"database/sql"
	"fmt"
)

func Connect() {
	db, err := sql.Open("postgres", getSqlInfo())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	createTaskTable :=
		`CREATE TABLE tasks(
			id SERIAL PRIMARY KEY, 
			uuid UUID, 
			start_at TIMESTAMP, 
			target_url TEXT, 
			interval_time INT, 
			interval_type VARCHAR(15), 
			send_at VARCHAR(10), 
			exec_type VARCHAR(10), 
			exec_body JSONB,
			exec_header JSONB, 
			exec_limit int, 
			immediately BOOL,
			continuous BOOL, 
			cancelled  BOOL, 
			fire_count INT, 
			successful_fire_count INT, 
			last_fire TIMESTAMP, 
			last_successful_fire TIMESTAMP)`
	_, err = db.Exec(createTaskTable)
	if err != nil {
		fmt.Println(err)
	}
}

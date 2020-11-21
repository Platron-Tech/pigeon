package db

import (
	"database/sql"
	"fmt"
)

func Connect() {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		user, password, host, port, dbname)

	//psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	//	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	createTaskTable := `CREATE TABLE tasks(id SERIAL PRIMARY KEY, uuid UUID, start_at TIMESTAMP, target_url TEXT, interval_time INT, interval_type VARCHAR(15), send_at VARCHAR(10), exec_type VARCHAR(10), exec_body JSONB,exec_header JSONB, exec_limit int, immediately BOOL,continuous BOOL, cancelled  BOOL, remain_count INT)`
	_, err = db.Exec(createTaskTable)
	if err != nil {
		fmt.Println(err)
	}
}

package db

import (
	"fmt"
	"os"
)

// if you run locally and does not have environment variables on your machine comment out this block
//func init() {
//	err := gotenv.Load(".env.sample")
//	if err != nil {
//		panic(err)
//	}
//}

func getSqlInfo() string {
	dbname := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_DB_USER")
	password := os.Getenv("POSTGRES_DB_PASS")
	host := os.Getenv("POSTGRES_DB_HOST")
	port := os.Getenv("POSTGRES_DB_PORT")

	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		user, password, host, port, dbname)
}

func GetApiKey() string {
	return os.Getenv("API_KEY")
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
)

type Logs struct {
	Message string
}

func main() {
	// welcome message
	fmt.Println(`
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
  Pipy pg-logger is running
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	`)
	//load ENV
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	dbAddr := os.Getenv("DB_ADDR")
	dbUser := os.Getenv("DB_USER")
	dbPasswd := os.Getenv("DB_PASSWD")
	dbName := os.Getenv("DB_NAME")
	svcListen := os.Getenv("SERVER_LISTENING")
	// gin-web
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	//db connect
	var pgsqlDB = pg.Connect(&pg.Options{
		Addr:     dbAddr,
		User:     dbUser,
		Password: dbPasswd,
		Database: dbName,
	})
	if pgsqlDB == nil {
		fmt.Println("error: pg.Connect() failed.")
		return
	}
	//db close
	defer func(pgsqlDB *pg.DB) {
		err := pgsqlDB.Close()
		if err != nil {
			fmt.Println("err: close postgresql failed.")
		}
	}(pgsqlDB)

	// save logs func
	r.POST("/logs", func(c *gin.Context) {
		fmt.Println("")
		fmt.Println("------------")
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Invalid request body")
		}
		// fmt.Println("【body】:  " + string(body))
		messages := strings.Split(string(body), "\n")

		var logList []Logs
		for _, v := range messages {
			// msg := json.NewDecoder(strings.NewReader(v))
			pipyLog := Logs{
				Message: v,
			}
			logList = append(logList, pipyLog)
		}

		result, err := pgsqlDB.Model(&logList).Insert()
		if err != nil {
			fmt.Println("batch insert rows error: ", err)
		} else {
			fmt.Printf("batch insert rows affected: %d\n", result.RowsAffected())
		}

		c.JSON(200, gin.H{
			"message": "success.",
		})
	})

	r.Run(svcListen)
}

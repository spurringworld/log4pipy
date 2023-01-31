package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
)

type Logs struct {
	Message   string
	ReqTime   uint64
	ResTime   uint64
	ReqPath   string
	ResStatus uint
}

type Message struct {
	Req struct {
		// Protocol string
		// Method   string
		Path string
	} `json:"req"`
	// ReqSize uint
	// ResSize uint
	Res struct {
		Status uint
	} `json:"res"`
	ReqTime uint64
	ResTime uint64
}

func main() {
	// welcome message
	fmt.Println(`
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~
            log4pipy is running    
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~
	`)
	// load ENV
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
	// db connect
	var pgDB = pg.Connect(&pg.Options{
		Addr:     dbAddr,
		User:     dbUser,
		Password: dbPasswd,
		Database: dbName,
	})
	if pgDB == nil {
		fmt.Println("error: pg.Connect() failed.")
		return
	}
	// db close
	defer func(pgsqlDB *pg.DB) {
		err := pgsqlDB.Close()
		if err != nil {
			fmt.Println("error: close postgresql failed.")
		}
	}(pgDB)

	// init gin-web
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// save logs func
	r.POST("/logs", func(c *gin.Context) {
		fmt.Println("")
		fmt.Println("------------")
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Invalid request body")
			c.JSON(406, gin.H{
				"error": "Invalid request body.",
			})
		}
		// fmt.Println("【body】:  " + string(body))
		messages := strings.Split(string(body), "\n")

		var logList []Logs
		for _, v := range messages {
			var msg Message
			json.Unmarshal([]byte(v), &msg)
			pipyLog := Logs{
				Message:   v,
				ReqTime:   msg.ReqTime,
				ResTime:   msg.ResTime,
				ReqPath:   msg.Req.Path,
				ResStatus: msg.Res.Status,
			}
			logList = append(logList, pipyLog)
		}

		result, err := pgDB.Model(&logList).Insert()
		if err != nil {
			fmt.Println("batch insert rows error: ", err)
		} else {
			fmt.Printf("batch insert rows affected: %d\n", result.RowsAffected())
		}

		c.JSON(200, gin.H{
			"message": "success.",
		})
	})

	// gin-web server run
	r.Run(svcListen)
}

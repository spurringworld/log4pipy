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
		var logList []Trafficlogs
		for _, v := range messages {
			var msg Message
			json.Unmarshal([]byte(v), &msg)
			pipyLog := Trafficlogs{
				Message:     v,
				ReqSize:     msg.ReqSize,
				ResSize:     msg.ResSize,
				ReqTime:     msg.ReqTime,
				ResTime:     msg.ResTime,
				EndTime:     msg.EndTime,
				RemoteAddr:  msg.RemoteAddr,
				LocalAddr:   msg.LocalAddr,
				RemotePort:  msg.RemotePort,
				LocalPort:   msg.LocalPort,
				ReqPath:     msg.Req.Path,
				ReqMethod:   msg.Req.Method,
				ResStatus:   msg.Res.Status,
				ServiceName: msg.Service.Name,
				MeshName:    msg.MeshName,
				ClusterName: msg.ClusterName,
				BondType:    msg.Type,
			}
			logList = append(logList, pipyLog)
		}
		result, err := pgDB.Model(&logList).Insert()
		if err != nil {
			fmt.Println("batch insert rows error: ", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			fmt.Printf("batch insert rows affected: %d\n", result.RowsAffected())
			c.JSON(200, gin.H{
				"message": "success.",
			})
		}
	})

	// gin-web server run
	r.Run(svcListen)
}

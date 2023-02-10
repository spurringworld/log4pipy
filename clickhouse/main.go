package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// welcome message
	fmt.Println(`
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
            log4pipy [clickhouse]   
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	`)
	// load ENV
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	dbServer := os.Getenv("DB_SERVER")
	dbUser := os.Getenv("DB_USER")
	dbPasswd := os.Getenv("DB_PASSWD")
	dbName := os.Getenv("DB_NAME")
	svcListen := os.Getenv("SERVER_LISTENING")
	// db connect
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dbServer},
		Auth: clickhouse.Auth{
			Database: dbName,
			Username: dbUser,
			Password: dbPasswd,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// init gin-web
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	///////////////////////
	// save logs func
	///////////////////////
	r.POST("/logs", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// fmt.Println("【body】:  " + string(body))
		messages := strings.Split(string(body), "\n")
		batch, err := conn.PrepareBatch(c, "INSERT INTO log (message)")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		for _, v := range messages {
			err := batch.Append(v)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}
		batch.Send()
		c.JSON(200, gin.H{"statusText": "success"})
	})

	///////////////////////
	// query logs func
	///////////////////////
	r.POST("/querylogs", func(c *gin.Context) {
		var logForm LogForm
		if err := c.ShouldBindJSON(&logForm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		baseSql := `
		SELECT service.name as ServiceName, pod.name as PodName, 
			req.path as ReqPath, req.method as ReqMethod, req.protocol as ReqProtocol,
			resTime as ResTime, reqTime as ReqTime, res.status as ResStatus, resSize as ResSize,
			remoteAddr as RemoteAddr, remotePort as RemotePort, localAddr as LocalAddr, localPort as LocalPort,
			timestamp as Timestamp, req.headers as ReqHeaders, message as Message
		FROM log
		WHERE  bondType != 'outbound'
		`
		customQuery := logForm.CustomQuery
		if len(customQuery) > 0 {
			baseSql += fmt.Sprintf(" AND (%s) ", customQuery)
		}
		reqTimeStart := logForm.ReqTimeStart
		if reqTimeStart > 0 {
			baseSql += fmt.Sprintf(" AND ReqTime > %d ", reqTimeStart)
		}
		reqTimeEnd := logForm.ReqTimeEnd
		if reqTimeEnd > 0 {
			baseSql += fmt.Sprintf(" AND ReqTime < %d ", reqTimeEnd)
		}
		// get total
		var total uint64
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s)", baseSql)
		// fmt.Println(countSql)
		row := conn.QueryRow(c, countSql)
		if err := row.Scan(&total); err != nil {
			fmt.Println("error: ", err)
		}
		// get data
		orderByField := "ReqTime"
		if len(logForm.OrderByField) > 0 {
			orderByField = logForm.OrderByField
		}
		orderByType := "desc"
		if len(logForm.OrderByType) > 0 {
			orderByType = logForm.OrderByType
		}
		limitStart := 0
		if logForm.LimitStart > 0 {
			limitStart = logForm.LimitStart
		}
		limitSize := 10
		if logForm.LimitSize > 0 {
			limitSize = logForm.LimitSize
		}
		querySql := baseSql + fmt.Sprintf(" ORDER BY %s %s LIMIT %d, %d", orderByField, orderByType, limitStart, limitSize)
		// fmt.Println(querySql)
		var result []struct {
			ServiceName string
			PodName     string
			ReqPath     string
			ReqMethod   string
			ReqProtocol string
			ResTime     uint64
			ReqTime     uint64
			ResStatus   uint32
			ResSize     uint64
			RemoteAddr  string
			RemotePort  uint32
			LocalAddr   string
			LocalPort   uint32
			Timestamp   time.Time
			ReqHeaders  string
			Message     string
		}
		if err := conn.Select(c, &result, querySql); err != nil {
			fmt.Println("error: ", err)
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      total,
		})
	})

	// gin-web server run
	r.Run(svcListen)
}

package clickhouse

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

func Run() {
	// welcome message
	fmt.Println(`
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
            log4pipy [clickhouse]   
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	`)
	// load ENV
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env file is missed")
		// return
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
		tableName := c.Query("table")
		if tableName == "" {
			tableName = "log"
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// fmt.Println("【body】:  " + string(body))
		messages := strings.Split(string(body), "\n")
		batch, err := conn.PrepareBatch(c, fmt.Sprintf("INSERT INTO %s (message)", tableName))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		for _, v := range messages {
			err := batch.Append(v)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		error3 := batch.Send()
		if error3 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": error3.Error()})
			return
		}
		c.JSON(200, gin.H{"statusText": "success"})
	})

	///////////////////////
	// to ping func
	///////////////////////
	r.GET("/logs", func(c *gin.Context) {
		baseSql := "SELECT NOW()"
		var result time.Time
		row := conn.QueryRow(c, baseSql)
		if err := row.Scan(&result); err != nil {
			fmt.Println("error: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
		})
	})

	///////////////////////
	// query logs func（for test）
	///////////////////////
	r.GET("/querylogs", func(c *gin.Context) {
		tableName := c.Query("table")
		if tableName == "" {
			tableName = "log"
		}
		baseSql := fmt.Sprintf(`
		SELECT service.name as ServiceName, pod.name as PodName, 
			req.path as ReqPath, req.method as ReqMethod, req.protocol as ReqProtocol,
			resTime as ResTime, reqTime as ReqTime, res.status as ResStatus, resSize as ResSize,
			remoteAddr as RemoteAddr, remotePort as RemotePort, localAddr as LocalAddr, localPort as LocalPort,
			timestamp as CreatedAt, req.headers as ReqHeaders, message as Message
		FROM %s
		WHERE  bondType != 'outbound'
		`, tableName)
		// get total
		var total uint64
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s)", baseSql)
		row := conn.QueryRow(c, countSql)
		if err := row.Scan(&total); err != nil {
			fmt.Println("error: ", err)
		}
		// get data
		querySql := baseSql + fmt.Sprintf(" LIMIT %d, %d", 0, 1)
		var result []LogVO
		if err := conn.Select(c, &result, querySql); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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

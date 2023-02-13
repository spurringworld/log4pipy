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
		reqTimeFrom := logForm.ReqTimeFrom
		if reqTimeFrom > 0 {
			baseSql += fmt.Sprintf(" AND ReqTime > %d ", reqTimeFrom)
		}
		reqTimeTo := logForm.ReqTimeTo
		if reqTimeTo > 0 {
			baseSql += fmt.Sprintf(" AND ReqTime < %d ", reqTimeTo)
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
		var result []Trafficlogs
		if err := conn.Select(c, &result, querySql); err != nil {
			fmt.Println("error: ", err)
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      total,
		})
	})

	///////////////////////
	// query service logs func
	///////////////////////
	r.POST("/querysvclogs", func(c *gin.Context) {
		var logForm SvcLogForm
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
		var whereSql = buildWhereSql(logForm)
		baseSql += whereSql

		// get total
		var total uint64
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s)", baseSql)
		row := conn.QueryRow(c, countSql)
		if err := row.Scan(&total); err != nil {
			fmt.Println("error: ", err)
		}
		// get data
		limitStart := 0
		if logForm.LimitStart > 0 {
			limitStart = logForm.LimitStart
		}
		limitSize := 10
		if logForm.LimitSize > 0 {
			limitSize = logForm.LimitSize
		}
		querySql := baseSql + fmt.Sprintf(" ORDER BY Timestamp desc LIMIT %d, %d", limitStart, limitSize)
		var result []Trafficlogs
		if err := conn.Select(c, &result, querySql); err != nil {
			fmt.Println("error: ", err)
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      total,
		})
	})

	///////////////////////
	// chart4latency func
	///////////////////////
	r.POST("/chart4latency", func(c *gin.Context) {
		var logForm SvcLogForm
		if err := c.ShouldBindJSON(&logForm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		baseSql := `
		SELECT (CEIL ((resTime - reqTime)/ 1000)) AS Latency, COUNT(1) as Count
		FROM log
		WHERE bondType != 'outbound'
		`
		var whereSql = buildWhereSql(logForm)
		querySql := baseSql + whereSql + " GROUP BY Latency ORDER BY Latency"
		// fmt.Println(querySql)
		var result []struct {
			Latency float64
			Count   uint64
		}
		if err := conn.Select(c, &result, querySql); err != nil {
			fmt.Println("error: ", err)
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
		})
	})

	// gin-web server run
	r.Run(svcListen)
}

func buildWhereSql(logForm SvcLogForm) string {
	var whereSql string
	svcName := logForm.ServiceName
	if len(svcName) > 0 {
		whereSql = fmt.Sprintf(" AND service.name = '%s' ", svcName)
	}
	queryWords := logForm.QueryWords
	if len(queryWords) > 0 {
		whereSql = whereSql + " AND message like '%" + queryWords + "%' "
	}
	reqTimeFrom := logForm.ReqTimeFrom //e.g. reqTimeFrom=15 day
	if len(reqTimeFrom) > 0 {
		whereSql += fmt.Sprintf(" AND toDateTime(reqTime / 1000) > now() - interval %s ", reqTimeFrom)
	}
	reqTimeTo := logForm.ReqTimeTo //e.g. reqTimeTo=1 second
	if len(reqTimeTo) > 0 {
		whereSql += fmt.Sprintf(" AND toDateTime(reqTime / 1000) < now() - interval %s ", reqTimeTo)
	}
	return whereSql
}

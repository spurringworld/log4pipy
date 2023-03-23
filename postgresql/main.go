package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
)

func main() {
	// welcome message
	fmt.Println(`
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
            log4pipy [postgresql]    
        ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
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
	var conn = pg.Connect(&pg.Options{
		Addr:     dbAddr,
		User:     dbUser,
		Password: dbPasswd,
		Database: dbName,
	})
	if conn == nil {
		fmt.Println("error: pg.Connect() failed.")
		return
	}
	// db close
	defer func(pgsqlDB *pg.DB) {
		err := pgsqlDB.Close()
		if err != nil {
			fmt.Println("error: close postgresql failed.")
		}
	}(conn)

	// init gin-web
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	///////////////////////
	// save logs func
	///////////////////////
	r.POST("/logs", func(c *gin.Context) {
		tableName := c.Query("table")
		if tableName == "" {
			tableName = "trafficlogs"
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Invalid request body")
			c.JSON(406, gin.H{
				"error": "Invalid request body.",
			})
		}
		// fmt.Println("【body】:  " + string(body))
		messages := strings.Split(string(body), "\n")
		baseSql := fmt.Sprintf(`
		  insert into %s (req_size, res_size, req_time, res_time, end_time, 
			remote_addr, local_addr, remote_port, local_port, req_path, req_method, 
			req_protocol, req_headers, res_status, service_name, pod_name, mesh_name,
			cluster_name, bond_type, message) values 
		`, tableName)
		var logList []Trafficlogs
		idx := 0
		for _, v := range messages {
			var msg Message
			json.Unmarshal([]byte(v), &msg)
			if idx > 0 {
				baseSql += ", "
			}
			valueSql := msg.ReqSize + ","
			valueSql += msg.ResSize + ","
			valueSql += msg.ReqTime + ","
			valueSql += msg.ResTime + ","
			valueSql += msg.EndTime + ","

			valueSql += msg.RemoteAddr + ","
			valueSql += msg.LocalAddr + ","
			valueSql += msg.RemotePort + ","
			valueSql += msg.LocalPort + ","

			valueSql += msg.Req.Path + ","
			valueSql += msg.Req.Method + ","
			valueSql += msg.Req.Protocol + ","
			valueSql += msg.Req.Headers + ","
			valueSql += msg.msg.Res.Status + ","
			valueSql += msg.Service.Name + ","
			valueSql += msg.Pod.Name + ","
			valueSql += msg.MeshName + ","
			valueSql += msg.ClusterName + ","
			valueSql += msg.Type + ","
			valueSql += v

			// pipyLog := Trafficlogs{
			// 	Message:     v,
			// 	ReqSize:     msg.ReqSize,
			// 	ResSize:     msg.ResSize,
			// 	ReqTime:     msg.ReqTime,
			// 	ResTime:     msg.ResTime,
			// 	EndTime:     msg.EndTime,
			// 	RemoteAddr:  msg.RemoteAddr,
			// 	LocalAddr:   msg.LocalAddr,
			// 	RemotePort:  msg.RemotePort,
			// 	LocalPort:   msg.LocalPort,
			// 	ReqPath:     msg.Req.Path,
			// 	ReqMethod:   msg.Req.Method,
			// 	ReqProtocol: msg.Req.Protocol,
			// 	ReqHeaders:  msg.Req.Headers,
			// 	ResStatus:   msg.Res.Status,
			// 	ServiceName: msg.Service.Name,
			// 	PodName:     msg.Pod.Name,
			// 	MeshName:    msg.MeshName,
			// 	ClusterName: msg.ClusterName,
			// 	BondType:    msg.Type,
			// }
			// logList = append(logList, pipyLog)

			baseSql += fmt.Sprintf("(%s)", valueSql)
			idx++
		}
		result, err := conn.Model(&logList).TableExpr(tableName).Insert()
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

	///////////////////////
	// to ping func
	///////////////////////
	r.GET("/logs", func(c *gin.Context) {
		baseSql := "SELECT NOW() AS nt"
		// get total
		var result struct {
			Nt time.Time
		}
		_, err1 := conn.QueryOne(&result, baseSql)
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
			return
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result.Nt,
		})
	})

	///////////////////////
	// query logs func(for test)
	///////////////////////
	r.GET("/querylogs", func(c *gin.Context) {
		tableName := c.Query("table")
		if tableName == "" {
			tableName = "trafficlogs"
		}
		baseSql := fmt.Sprintf(`
		SELECT service_name, pod_name, req_path, req_method, req_protocol,
			res_time, req_time, res_status, res_size,
			remote_addr, remote_port, local_addr, local_port,
			created_at, req_headers, message
		FROM %s
		WHERE  bond_type != 'outbotrafficlogsund'
		`, tableName)
		// get total
		var countResult struct {
			Total uint64
		}
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s) abc", baseSql)
		// fmt.Println(countSql)
		_, err1 := conn.QueryOne(&countResult, countSql)
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
			return
		}
		// get data
		querySql := baseSql + " LIMIT 1 OFFSET 0"
		// fmt.Println(querySql)
		var result []LogVO
		_, err2 := conn.Query(&result, querySql)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      countResult.Total,
		})
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
		SELECT service_name, pod_name, req_path, req_method, req_protocol,
			res_time, req_time, res_status, res_size,
			remote_addr, remote_port, local_addr, local_port,
			created_at, req_headers, message
		FROM trafficlogs
		WHERE  bond_type != 'outbound'
		`
		customQuery := logForm.CustomQuery
		if len(customQuery) > 0 {
			baseSql += fmt.Sprintf(" AND (%s) ", customQuery)
		}
		reqTimeFrom := logForm.ReqTimeFrom
		if reqTimeFrom > 0 {
			baseSql += fmt.Sprintf(" AND req_time > %d ", reqTimeFrom)
		}
		reqTimeTo := logForm.ReqTimeTo
		if reqTimeTo > 0 {
			baseSql += fmt.Sprintf(" AND req_time < %d ", reqTimeTo)
		}
		// get total
		var countResult struct {
			Total uint64
		}
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s) abc", baseSql)
		// fmt.Println(countSql)
		_, err1 := conn.QueryOne(&countResult, countSql)
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
			return
		}
		// get data
		orderByField := "req_time"
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
		querySql := baseSql + fmt.Sprintf(" ORDER BY %s %s LIMIT %d OFFSET %d", orderByField, orderByType, limitSize, limitStart)
		// fmt.Println(querySql)
		var result []LogVO
		_, err2 := conn.Query(&result, querySql)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      countResult.Total,
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
		SELECT service_name, pod_name, req_path, req_method, req_protocol,
			res_time, req_time, res_status, res_size,
			remote_addr, remote_port, local_addr, local_port,
			created_at, req_headers, message
		FROM trafficlogs
		WHERE  bond_type != 'outbound'
		`
		var whereSql = buildWhereSql(logForm)
		baseSql += whereSql

		// get total
		var countResult struct {
			Total uint64
		}
		countSql := fmt.Sprintf("SELECT count(1) AS total FROM (%s) abc", baseSql)
		// fmt.Println(countSql)
		_, err1 := conn.QueryOne(&countResult, countSql)
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
			return
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
		querySql := baseSql + fmt.Sprintf(" ORDER BY created_at desc LIMIT %d OFFSET %d", limitSize, limitStart)
		// fmt.Println(querySql)
		var result []LogVO
		_, err2 := conn.Query(&result, querySql)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}
		c.JSON(200, gin.H{
			"statusText": "success",
			"data":       result,
			"total":      countResult.Total,
		})
	})

	// gin-web server run
	r.Run(svcListen)
}

//////////////////////////////
// common method buildWhereSql
//////////////////////////////
func buildWhereSql(logForm SvcLogForm) string {
	var whereSql string
	svcName := logForm.ServiceName
	if len(svcName) > 0 {
		whereSql = fmt.Sprintf(" AND service_name = '%s' ", svcName)
	}
	queryWords := logForm.QueryWords
	if len(queryWords) > 0 {
		whereSql = whereSql + " AND cast(message AS varchar) like '%" + queryWords + "%' "
	}
	reqTimeFrom := logForm.ReqTimeFrom //e.g. reqTimeFrom=15 day
	if len(reqTimeFrom) > 0 {
		whereSql += fmt.Sprintf(" AND to_timestamp(req_time / 1000) > now() - interval '%s' ", reqTimeFrom)
	}
	reqTimeTo := logForm.ReqTimeTo //e.g. reqTimeTo=1 second
	if len(reqTimeTo) > 0 {
		whereSql += fmt.Sprintf(" AND to_timestamp(req_time / 1000) < now() - interval '%s' ", reqTimeTo)
	}
	return whereSql
}

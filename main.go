package main

import (
	rateClient "ddl-api/client"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/hduLib/hdu/client"
	"net/http"
	"strconv"
	"sync"
)

const threads = 64

type DDL struct {
	Course string `json:"course"`
	Title  string `json:"title"`
	Time   int64  `json:"time"`
	Type   string `json:"type"`
	From   string `json:"from"`
}

type RespForm struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

var port int
var lock = new(sync.Mutex)
var wg sync.WaitGroup

func getAndAppend(ddls *[]DDL, errs *[]error, f func() ([]DDL, []error)) {
	d, e := f()
	lock.Lock()
	defer func() {
		lock.Unlock()
		wg.Done()
	}()
	*ddls = append(*ddls, d...)
	*errs = append(*errs, e...)
}

func handleRequest(c *gin.Context) {
	var ddls []DDL
	var errs []error
	if c.Query("cx_account") != "" {
		wg.Add(1)
		go getAndAppend(&ddls, &errs, func() ([]DDL, []error) {
			return GettingCxddl(c.Query("cx_account"), c.Query("cx_passwd"), c.Query("cx_loginType"))
		})
	}
	if c.Query("zjooc_account") != "" {
		wg.Add(1)
		go getAndAppend(&ddls, &errs, func() ([]DDL, []error) {
			return GettingZjoocDdl(c.Query("zjooc_account"), c.Query("zjooc_passwd"))
		})
	}
	wg.Wait()
	if errs != nil {
		c.JSON(http.StatusBadRequest, &RespForm{
			Code:   -1,
			Data:   ddls,
			Errors: errs2strs(errs),
		})
		return
	}
	c.JSON(http.StatusOK, &RespForm{
		Code: 0,
		Data: ddls,
		Msg:  "ok",
	})
}

func main() {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()
	client.DefaultClient = rateClient.NewWaitingClient(threads)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ddl/all", handleRequest)
	r.Run(":" + strconv.Itoa(port))
}

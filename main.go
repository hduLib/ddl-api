package main

import (
	"ddl-api/client"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/hduLib/hdu/net"
	"net/http"
	"strconv"
)

const threads = 32

type ddl struct {
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

func handleRequest(c *gin.Context) {
	ddls, errs := GettingCxddl(c.Query("cx_account"), c.Query("cx_passwd"), c.Query("cx_loginType"))
	if errs != nil {
		c.JSON(http.StatusBadRequest, &RespForm{
			Code:   -1,
			Data:   nil,
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
	net.DefaultClient = client.NewWaitingClient(threads)
	r := gin.Default()
	r.GET("/ddl/all", handleRequest)
	r.Run(":" + strconv.Itoa(port))
}

package main

import (
	rateClient "ddl-api/client"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/hduLib/hdu/client"
	"github.com/hduLib/hdu/skl"
	"github.com/hduLib/hdu/skl/schema"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const threads = 64

type DDL struct {
	Course string `json:"course"`
	Title  string `json:"title"`
	Time   int64  `json:"time"`
	Type   string `json:"type"`
	From   string `json:"from"`
}
type Course struct {
	CourseName   string `json:"courseName"`
	StartSection int    `json:"startSection"`
	EndSection   int    `json:"endSection"`
	ClassRoom    string `json:"classRoom"`
	WeekDay      int    `json:"weekDay"`
	TeacherName  string `json:"teacherName"`
	CourseCode   string `json:"courseCode"`
	CourseType   string `json:"courseType"`
}

func NewCourse(c skl.Course) *Course {
	return &Course{
		CourseName:   c.CourseName,
		StartSection: c.StartSection,
		EndSection:   c.EndSection,
		ClassRoom:    c.ClassRoom,
		WeekDay:      c.WeekDay,
		TeacherName:  c.TeacherName,
		CourseCode:   c.CourseCode,
		CourseType:   c.CourseType,
	}
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

func handleDDLRequest(c *gin.Context) {
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

func parseByDay(s schema.Schema, wd time.Weekday, wn int) (scheduled [][2]int) {
	for _, v := range s {
		if v.Weekday == wd {
			if v.WeekNum.Check(wn) {
				scheduled = append(scheduled, [2]int{v.Begin, v.End})
			}
		}
	}
	return
}

func handleCourseRequest(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, &RespForm{
			Code: -1,
			Msg:  "用户名或密码为空",
		})
		return
	}
	user, err := skl.Login(username, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, &RespForm{
			Code: -1,
			Msg:  "用户名或密码错误",
		})
		return
	}
	resp, err := user.Course(time.Now())
	if err != nil {
		r := rand.Int31()
		log.Printf("%v(%d)\n", err, r)
		c.JSON(http.StatusBadRequest, &RespForm{
			Code: -1,
			Msg:  "获取课表失败(" + strconv.FormatInt(int64(r), 32) + ")",
		})
		return
	}
	var list []*Course
	for _, v := range resp.List {
		decode, err := schema.Decode(v.CourseSchema)
		if err != nil {
			log.Println("fail to decode schema:", v.CourseSchema)
		}
		schedule := parseByDay(decode, time.Now().Weekday(), resp.Week)
		if schedule != nil {
			for _, w := range schedule {
				existed := false
				for _, x := range list {
					if x.StartSection == w[0] {
						existed = true
						break
					}
				}
				if !existed {
					c := NewCourse(v)
					c.StartSection, c.EndSection = w[0], w[1]
					list = append(list, c)
				}
			}
		}
	}
	c.JSON(http.StatusOK, &RespForm{
		Code: 0,
		Msg:  "ok",
		Data: struct {
			Week      int       `json:"week"`
			Xn        string    `json:"xn"`
			Xq        string    `json:"xq"`
			StartTime int64     `json:"startTime"`
			List      []*Course `json:"list"`
		}{
			Week:      resp.Week,
			Xn:        resp.Xn,
			Xq:        resp.Xq,
			StartTime: resp.StartTime,
			List:      list,
		},
	})
}

func main() {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()
	client.DefaultClient = rateClient.NewWaitingClient(threads)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ddl/all", handleDDLRequest)
	r.GET("/courses/today", handleCourseRequest)
	r.Run(":" + strconv.Itoa(port))
}

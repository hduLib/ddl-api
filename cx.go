package main

import (
	"ddl-api/collector"
	"errors"
	"fmt"
	cx "github.com/hduLib/hdu/chaoxing"
	"sync"
	"time"
)

func GettingCxddl(account, passwd, Type string) ([]DDL, []error) {
	var user *cx.Cx
	var err error
	if Type == "cx" {
		user, err = cx.LoginWithPhoneAndPwd(account, passwd)
		if err != nil {
			return nil, []error{fmt.Errorf("cx login fail:%v", err)}
		}
	} else if Type == "cas" {
		user, err = cx.LoginWithCas(account, passwd)
		if err != nil {
			return nil, []error{fmt.Errorf("cas login fail:%v", err)}
		}
	} else {
		return nil, []error{errors.New("wrong cx login type")}
	}

	list, err := user.CourseList()
	if err != nil {
		return nil, []error{fmt.Errorf("get course list fail:%v", err)}
	}
	collect := collector.New[DDL]()
	wg := &sync.WaitGroup{}
	for _, course := range list.Courses {
		course := course
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := course.Detail()
			if err != nil {
				collect.ErrCollector <- fmt.Errorf("fail to get course detail:%v", err)
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				examList, err := c.ExamList()
				if err != nil {
					collect.ErrCollector <- fmt.Errorf("fail to get exam list:%v", err)
					return
				}
				for _, v := range examList.Exams {
					if v.Status == "待做" {
						collect.Add <- DDL{
							course.Title, v.Title, v.Time.Unix(), "考试", "超星",
						}
					}
				}
			}()
		}()
	}
	wkls, err := user.WorkList()
	if err != nil {
		ddl, err := collect.Done()
		return ddl, []error{fmt.Errorf("fail to get work list:%v", err)}
	}
	t := time.Now()
	for _, work := range wkls.Works {
		if work.Time.After(t) && work.Status == "未提交" {
			var title string
			for _, v := range list.Courses {
				if v.ClazzId == work.ClazzId {
					title = v.Title
				}
			}
			collect.Add <- DDL{
				title, work.Title, work.Time.Unix(), "作业", "超星",
			}
		}
	}
	wg.Wait()
	return collect.Done()
}

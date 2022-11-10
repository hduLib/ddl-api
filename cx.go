package main

import (
	"errors"
	"fmt"
	cx "github.com/hduLib/hdu/chaoxing"
	"github.com/hduLib/hdu/client"
	"sync"
)

func GettingCxddl(account, passwd, Type string) ([]ddl, []error) {
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

	var todo []ddl
	errCollector := make(chan error, 8)
	add := make(chan ddl, 8)
	wg := &sync.WaitGroup{}
	for _, course := range list.Courses {
		course := course
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := course.Detail()
			if err != nil {
				if err, ok := err.(*client.ErrNotOk); ok {
					errCollector <- fmt.Errorf("fail to get course detail:status code %d: %s", err.StatusCode, err.Body)
					return
				}
				errCollector <- fmt.Errorf("fail to get course detail:%v", err)
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				workList, err := c.WorkList()
				if err != nil {
					if err, ok := err.(*client.ErrNotOk); ok {
						errCollector <- fmt.Errorf("fail to get work list:status code %d: %s", err.StatusCode, err.Body)
						return
					}
					errCollector <- fmt.Errorf("fail to get work list:%v", err)
					return
				}
				for _, v := range workList.Works {
					if v.Status == "未交" && v.Time.Unix() != 0 {
						add <- ddl{
							course.Title, v.Title, v.Time.Unix(), "作业", "超星",
						}
					}
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				examList, err := c.ExamList()
				if err != nil {
					errCollector <- fmt.Errorf("fail to get exam list:%v", err)
					return
				}
				for _, v := range examList.Exams {
					if v.Status == "待做" {
						add <- ddl{
							course.Title, v.Title, v.Time.Unix(), "考试", "超星",
						}
					}
				}
			}()
		}()
	}
	// collector
	Done := make(chan struct{})
	go func() {
		for {
			select {
			case <-Done:
				return
			case a := <-add:
				todo = append(todo, a)
			}
		}
	}()
	// fmt.Println("tasks:", tasks)
	var errs []error
	go func() {
		for {
			select {
			case err := <-errCollector:
				if err != nil {
					errs = append(errs, err)
				}
			case <-Done:
				return
			}
		}
	}()
	wg.Wait()
	// close 2 goroutine
	Done <- struct{}{}
	Done <- struct{}{}
	if errs != nil {
		return nil, errs
	}
	return todo, nil
}

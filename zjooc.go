package main

import (
	"fmt"
	"github.com/hduLib/hdu/zjooc"
	"sync"
	"time"
)

func GettingZjoocDdl(account, passwd string) ([]ddl, []error) {
	user, err := zjooc.Login(account, passwd)
	if err != nil {
		return nil, []error{fmt.Errorf("zjooc login fail:%v", err)}
	}
	list, err := user.CurrentCourses(zjooc.Published)
	if err != nil {
		return nil, []error{fmt.Errorf("get zjooc course list fail:%v", err)}
	}

	var todo []ddl
	errCollector := make(chan error, 8)
	add := make(chan ddl, 8)
	wg := &sync.WaitGroup{}
	for _, course := range list {
		course := course
		find := func(Type int, typename string) {
			defer wg.Done()
			workList, err := user.PapersByCourse(course.Id, Type, course.BatchId)
			if err != nil {
				errCollector <- fmt.Errorf("fail to get zjooc assignment list:%v", err)
				return
			}
			for _, v := range workList {
				if v.ReviewStatus == 0 && v.ProcessStatus == 0 {
					t, err := time.Parse("2006-01-02T15:04:05.000-0700", v.EndTime)
					if err != nil {
						errCollector <- fmt.Errorf("fail to get zjooc assignment:%v", err)
						return
					}
					add <- ddl{
						v.CourseName, v.PaperName, t.Unix(), typename, "在浙学",
					}
				}
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			wg.Add(1)
			find(zjooc.Assignment, "作业")
			wg.Add(1)
			find(zjooc.Test, "测验")
			wg.Add(1)
			find(zjooc.Exam, "考试")
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

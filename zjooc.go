package main

import (
	"ddl-api/collector"
	"fmt"
	"github.com/hduLib/hdu/zjooc"
	"sync"
	"time"
)

func GettingZjoocDdl(account, passwd string) ([]DDL, []error) {
	user, err := zjooc.Login(account, passwd)
	if err != nil {
		return nil, []error{fmt.Errorf("zjooc login fail:%v", err)}
	}
	list, err := user.CurrentCourses(zjooc.Published)
	if err != nil {
		return nil, []error{fmt.Errorf("get zjooc course list fail:%v", err)}
	}

	collect := collector.New[DDL]()
	wg := &sync.WaitGroup{}
	for _, course := range list {
		course := course
		find := func(Type int, typename string) {
			defer wg.Done()
			workList, err := user.PapersByCourse(course.Id, Type, course.BatchId)
			if err != nil {
				collect.ErrCollector <- fmt.Errorf("fail to get zjooc assignment list:%v", err)
				return
			}
			for _, v := range workList {
				if v.ReviewStatus == 0 && v.ProcessStatus == 0 {
					t, err := time.Parse("2006-01-02T15:04:05.000-0700", v.EndTime)
					if err != nil {
						collect.ErrCollector <- fmt.Errorf("fail to get zjooc assignment:%v", err)
						return
					}
					collect.Add <- DDL{
						v.CourseName, v.PaperName, t.Unix(), typename, "在浙学",
					}
				}
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			wg.Add(3)
			find(zjooc.Assignment, "作业")
			find(zjooc.Test, "测验")
			find(zjooc.Exam, "考试")
		}()
	}
	wg.Wait()
	return collect.Done()
}

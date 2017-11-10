package sample

import (
	"fmt"
	"log"
	"time"

	jira "gopkg.in/andygrunwald/go-jira.v1"
)

const (
	Span = 6
)

func GetCounter() []Record {
	issues, e := getRecords()
	if e != nil {
		return nil
	}
	v := assortIssues(issues)
	return ExchangeRecord(v)
}
func ExchangeRecord(mp map[string][]int) []Record {
	rec := make([]Record, 0, 100)
	for i, v := range mp {
		obj := Record{
			Name:    i,
			Counter: v,
		}
		rec = append(rec, obj)
	}
	return rec
}

type Record struct {
	Name    string
	Counter []int
}

type Result struct {
	Record map[string][]int
}

func assortIssues(vals []jira.Issue) map[string][]int {
	ret := make(map[string][]int)

	for _, v := range vals {
		name := createName(v.Fields.Assignee)
		dayIndex := createDayIndex(v.Fields.Resolutiondate)
		if _, ok := ret[name]; ok == false {
			ret[name] = make([]int, Span)
			ret[name][dayIndex] = 1
		} else {
			ret[name][dayIndex] += 1
		}
	}
	return ret
}
func createDayIndex(date string) int {
	t, e := time.Parse("2006-01-02T15:04:05.000+0000", date)
	if e != nil {
		return 0
	}
	duration := int((time.Now().Sub(t)).Hours())
	if duration > 24 {
		result := (int(duration) / 24) / 7
		return result
	}
	return 0
}
func createName(user *jira.User) string {
	if user != nil {
		return user.Name
	}
	return "empty"
}
func getRecords() ([]jira.Issue, error) {
	//	Jiraとの接続
	jiraClient, _ := jira.NewClient(nil, "http://ec2-34-211-174-71.us-west-2.compute.amazonaws.com/jira/")
	jiraClient.Authentication.SetBasicAuth("support_asmil", "rzXPJb6V")

	//　課題の取得
	opt := &jira.SearchOptions{MaxResults: 1000}

	issues, _, err := jiraClient.Issue.Search(fmt.Sprintf("resolutiondate >= -%dw ORDER BY assignee ASC, updated DESC", Span), opt)
	if err == nil {
		log.Printf("cnt:%d\n", len(issues))
		return issues, nil
	}
	return nil, err

}

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/kaepa3/jirawatcher/sample"

	"github.com/BurntSushi/toml"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

var config Config

type Config struct {
	JiraURL string
	User    string
	Pass    string
	Jql     string
}

func main() {
	initialize()
	goji.Get("/", watchPage)
	goji.Serve()
}

func initialize() {
	toml.DecodeFile("./config.toml", &config)
}

var mainTmpl *template.Template = template.Must(template.ParseFiles("tmpl/main.tmpl"))

func watchPage(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "text/html: charset=utf-8")
	vals, err := getIssues()
	if err == nil {
		counter := sample.GetCounter()
		if counter != nil {
			var vm ViewModel
			vm.Records = assortIssues(vals)
			vm.Url = config.JiraURL
			vm.GraphHeader, vm.GraphValue = createTemplateText(counter)
			mainTmpl.Execute(w, vm)
		}
	} else {
		fmt.Fprintf(w, err.Error())
	}
}
func createTemplateText(counter []sample.Record) ([]string, [][]int) {

	header := make([]string, 0, 10)
	for _, v := range counter {
		header = append(header, v.Name)
	}

	total := make([][]int, 0, 10)
	for i := range counter[0].Counter {
		data := make([]int, 0, 10)
		for _, v := range counter {
			data = append(data, v.Counter[i])
		}
		total = append(total, data)
	}
	return header, total
}

type ViewModel struct {
	Records     map[string][]Result
	GraphHeader []string
	GraphValue  [][]int
	Url         string
}

type Result struct {
	Title    string
	Key      string
	Project  string
	Assignee string
	Status   string
}

func assortIssues(vals []jira.Issue) map[string][]Result {
	ret := make(map[string][]Result)
	for _, v := range vals {
		if _, ok := ret[v.Fields.Assignee.Name]; ok == false {
			ret[v.Fields.Assignee.Name] = make([]Result, 0, 100)
		}
		ret[v.Fields.Assignee.Name] = append(ret[v.Fields.Assignee.Name], Result{
			v.Fields.Summary,
			v.Key,
			v.Fields.Project.Name,
			v.Fields.Assignee.Name,
			v.Fields.Status.Name,
		})
	}
	return ret
}

func getIssues() ([]jira.Issue, error) {
	//	Jiraとの接続
	jiraClient, _ := jira.NewClient(nil, config.JiraURL)
	jiraClient.Authentication.SetBasicAuth(config.User, config.Pass)

	//　課題の取得
	opt := &jira.SearchOptions{MaxResults: 1000}
	issues, _, err := jiraClient.Issue.Search(config.Jql, opt)
	if err == nil {
		log.Println("cnt:", len(issues))
		return issues, nil
	}
	return nil, err
}

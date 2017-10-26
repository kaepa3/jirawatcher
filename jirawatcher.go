package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

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
		mainTmpl.Execute(w, assortIssues(vals))
	} else {
		fmt.Fprintf(w, err.Error())
	}
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

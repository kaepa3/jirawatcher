package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/kaepa3/jirawatcher/sample"
	"github.com/kaepa3/jirawatcher/userauth"
	"github.com/kaepa3/oauth/lib"

	"github.com/BurntSushi/toml"
	log "github.com/cihub/seelog"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	v2 "google.golang.org/api/oauth2/v2"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

func initialize() {
	initLogger()
	flag.Set("bind", ":50000")
	toml.DecodeFile("./jiraConfig.toml", &config)
	auth = userauth.NewUserAuth("userauth.toml")
}

func initLogger() {
	logConfig := `
	<seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
		<formats>
		    <format id="main" format="Time:%Date(2006/01/02) %Time	file:%File	func:%FuncShort	line:%Line	level:%LEV	msg:%Msg%n" />
		    <format id="con" format="%Msg%n" />
		</formats>
		<outputs formatid="main">
			<rollingfile filename="rev.log" type="size" maxsize="102400" maxrolls="1" formatid = "main"/>
			<console formatid = "con"/>
		</outputs>
	</seelog>`
	logger, err := log.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	log.ReplaceLogger(logger)
}

var config JiraConfig
var auth *userauth.UserAuth

type JiraConfig struct {
	JiraURL string
	User    string
	Pass    string
	Jql     string
}

func main() {
	initialize()
	goji.Get("/", indexPage)
	goji.Get("/callback", watchPage)
	goji.Serve()
}

func indexPage(c web.C, w http.ResponseWriter, r *http.Request) {
	oauthConfig := google.GetConnect()
	url := oauthConfig.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusFound)
}

func watchPage(c web.C, w http.ResponseWriter, r *http.Request) {
	tokenInfo := createToken(r)
	if tokenInfo != nil {
		buf, err := tokenInfo.MarshalJSON()
		if err != nil {
			if auth.Authentication(tokenInfo) {
				displayInfomation(c, w, r)
			} else {
				fmt.Fprintf(w, string(buf))
			}
		}
	}
}
func createToken(r *http.Request) *v2.Tokeninfo {
	oauthConfig := google.GetConnect()

	context := context.Background()

	token, err := oauthConfig.Exchange(context, createCode(r))
	if err != nil {
		log.Info(err)
		return nil
	}

	if token.Valid() == false {
		log.Info("vaild token")
	}
	service, _ := v2.New(oauthConfig.Client(context, token))
	tokenInfo, _ := service.Tokeninfo().AccessToken(token.AccessToken).Context(context).Do()
	return tokenInfo
}

var mainTmpl *template.Template = template.Must(template.ParseFiles("tmpl/main.tmpl"))

func displayInfomation(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "text/html: charset=utf-8")
	vals, err := getIssues()
	if err == nil {
		counter := sample.GetCounter(config.JiraURL, config.User, config.Pass)
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
		log.Info("cnt:", len(issues))
		return issues, nil
	}
	return nil, err
}

func createCode(r *http.Request) string {
	for key, values := range r.URL.Query() {
		if key == "code" {
			for _, v := range values {
				return v
			}
		}
	}
	return ""
}

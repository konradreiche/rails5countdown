package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const GITHUB_API_ENDPOINT = "https://api.github.com/repos/rails/rails/issues?milestone=34&state=all&per_page=100&page=%d"

type Issue struct {
	State       string
	CreatedAt   *time.Time      `json:"created_at"`
	ClosedAt    *time.Time      `json:"closed_at"`
	PullRequest json.RawMessage `json:"pull_request"`
}

type Page struct {
	Digits []string
	Days   int
}

var daysLeft int = 0

func computeDaysLeft() {
	fmt.Println("Initialize loop to compute days left")
	for {
		fmt.Println("Update days left")
		daysLeft = estimateRelease(fetchIssues())
		s := fmt.Sprintf("Still %d days left", daysLeft)
		fmt.Println(s)
		fmt.Println("Going back to sleep")
		time.Sleep(1 * time.Hour)
	}
}

func fetchIssues() []Issue {
	var result []Issue

	for i := 1; i <= 3; i++ {
		url := fmt.Sprintf(GITHUB_API_ENDPOINT, i)
		res, err := http.Get(url)
		perror(err)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		perror(err)

		var issues []Issue
		err = json.Unmarshal(body, &issues)
		perror(err)

		result = append(result, issues...)
	}

	return result
}

func estimateRelease(issues []Issue) int {
	t, err := time.Parse("02-01-2006", "01-08-2015")
	sum := time.Duration(0)
	perror(err)

	closed := 0
	open := 0
	pullRequests := 0

	for _, issue := range issues {
		if issue.PullRequest == nil {
			if issue.State == "open" {
				open += 1
			} else if issue.CreatedAt.After(t) {
				delta := issue.ClosedAt.Sub(*issue.CreatedAt)
				sum += delta
				closed += 1
			}
		} else if issue.State == "closed" {
			pullRequests += 1
		}
	}

	return int(sum.Hours())/closed*open/24 + pullRequests
}

func perror(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	days := strings.Split(strconv.Itoa(daysLeft), "")
	page := &Page{Days: daysLeft, Digits: days}
	t, err := template.ParseFiles("index.html")
	perror(err)
	t.Execute(w, page)
}

func main() {
	go computeDaysLeft()

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9090", nil)
}

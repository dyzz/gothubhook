package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/fatih/structs"
	"github.com/hoisie/mustache"
)

type Event struct {
	Author string
	Repo   string
	Branch string
	Action string
	Msg    string
	Date   string
	Type   string
}

func NewEvent(author, repo, branch, action, msg, date, ty string) *Event {
	return &Event{
		Author: author,
		Repo:   repo,
		Branch: branch,
		Action: action,
		Msg:    msg,
		Date:   date,
		Type:   ty,
	}
}

var pushTmpl, _ = mustache.ParseString("{{Date}} -- {{Author}} {{Action}} to {{Repo}}/{{Branch}}\n\t{{Msg}}\n")
var pqTmpl, _ = mustache.ParseString("{{Date}} -- {{Author}} {{Action}} Pull Request{{Branch}} to {{Repo}}\n\t{{Msg}}\n")

func (e *Event) String() string {
	m := structs.Map(e)
	if e.Type == "push" {
		return pushTmpl.Render(m)
	}
	if e.Type == "pullrequest" {
		return pqTmpl.Render(m)
	}
	return ""
}

type Server struct {
	Port   int
	Path   string
	Secret string
	Events chan Event
}

func NewServer() *Server {
	return &Server{
		Port:   80,
		Path:   "/hook",
		Events: make(chan Event),
	}
}

func (srv *Server) Listen() error {
	return http.ListenAndServe(":"+strconv.Itoa(srv.Port), srv)
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
		msg := "Github hook request is POST only"
		http.Error(w, msg, http.StatusMethodNotAllowed)
		fmt.Println(msg)
		return
	}

	if req.URL.Path != srv.Path {
		msg := "Invalid Path"
		http.Error(w, msg, http.StatusNotFound)
		fmt.Println(msg)
		return
	}

	eventType := req.Header.Get("X-GitHub-Event")
	if eventType != "push" && eventType != "pull_request" {
		w.Write([]byte(fmt.Sprintf("Evernt Type %s is not supported", eventType)))
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Fail to read req.Body", http.StatusBadRequest)
	}

	if srv.Secret != "" {
		sig := req.Header.Get("X-Hub-Signature")

		if sig == "" {
			http.Error(w, "Missing X-Hub-Signature", http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(srv.Secret))
		mac.Write(body)
		expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
		// hmac Equal won't leak through timing side-channel
		if !hmac.Equal([]byte(expected), []byte(sig)) {
			http.Error(w, "Secret not match", http.StatusForbidden)
			return
		}
	}

	request, err := jason.NewObjectFromBytes(body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// https://developer.github.com/v3/activity/events/types/#pushevent
	// (*jason.Object)(0xc20803b590)({"added":["test.go"],"author":{"email":"naituida@163.com","name":"naituida","username":"dyzz"},"committer":{"email":"naituida@163.com","name":"naituida","username":"dyzz"},"distinct":true,"id":"76fa79bab843450020bd73fbba0b30741f99055f","message":"init push","modified":[],"removed":[],"timestamp":"2015-09-11T02:22:02-04:00","url":"https://github.com/dyzz/gothubhook/commit/76fa79bab843450020bd73fbba0b30741f99055f"}) }

	if eventType == "push" {
		repo, _ := request.GetString("repository", "name")
		ref, _ := request.GetString("ref")
		chunks := strings.Split(ref, "/")
		branch := chunks[len(chunks)-1]

		commits, _ := request.GetObjectArray("commits")
		for _, commit := range commits {
			author, _ := commit.GetString("author", "name")
			msg, _ := commit.GetString("message")
			date, _ := commit.GetString("timestamp")
			event := NewEvent(author, repo, branch, "commit", msg, date, "push")
			w.Write([]byte(event.String()))
			fmt.Println(event.String())
		}
		return
	}

	// https://developer.github.com/v3/activity/events/types/#pullrequestevent
	if eventType == "pull_request" {
		action, _ := request.GetString("action")
		number, _ := request.GetInt64("number")
		pq, _ := request.GetObject("pull_request")
		msg, _ := pq.GetString("title")
		date, _ := pq.GetString("updated_at")
		author, _ := pq.GetString("user", "login")
		repo, _ := request.GetString("repository", "name")
		event := NewEvent(author, repo, "#"+strconv.Itoa(int(number)), action, msg, date, "pullrequest")
		w.Write([]byte(event.String()))
		fmt.Println(event.String())
	}
}

func main() {
	srv := NewServer()
	err := srv.Listen()

	if err != nil {
		fmt.Println(err)
	}

	for {
		var output string
		fmt.Scanln(&output)
		fmt.Println(output)
	}

}

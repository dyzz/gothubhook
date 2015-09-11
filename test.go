package main

import (
	"fmt"
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
	Msg    string
	Type   string
}

func NewEvernt(author, repo, branch, msg, ty string) *Event {
	return &Event{
		Author: author,
		Repo:   repo,
		Branch: branch,
		Msg:    msg,
		Type:   ty,
	}
}

var eventTmpl, err = mustache.ParseString("<ul><li>{{Author}}</li><li>{{Repo}}</li><li>{{Branch}}</li><li>{{Msg}}</li></ul>")

func (e *Event) String() string {
	m := structs.Map(e)
	return eventTmpl.Render(m)
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

	request, err := jason.NewObjectFromReader(req.Body)
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
			event := NewEvernt(author, repo, branch, msg, "push")
			w.Write([]byte(event.String()))
			fmt.Println(event.String())
		}
		return
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

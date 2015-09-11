package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/antonholmquist/jason"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/structs"
	"github.com/hoisie/mustache"
)

type Event struct {
	Author string
	Repo   string
	Branch string
	Commit string
	Type   string
}

var eventTmpl, err = mustache.ParseString("<ul><li>{{Author}}</li><li>{{Repo}}</li><li>{{Branch}}</li><li>{{Commit}}</li></ul>")

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
	if eventType == "push" {
		commits, _ := request.GetObjectArray("commits")
		w.Write([]byte(spew.Sdump(commits)))
		spew.Dump(commits)
		return
	}
}

func main() {
	event := &Event{
		Repo:   "evm",
		Branch: "review",
		Commit: "fix bugs",
	}
	fmt.Println(event.String())
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

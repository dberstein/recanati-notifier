package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dberstein/recanati-notifier/delivery"
	httplog "github.com/dberstein/recanati-notifier/httplog"
	"github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/queue"
	"github.com/dberstein/recanati-notifier/user"

	"github.com/fatih/color"
)

var users []user.User

func init() {
	users = []user.User{
		{Id: 1, Mediums: []medium.Medium{medium.NewEmail("example@example.com")}},
		{Id: 2, Mediums: []medium.Medium{medium.NewEmail("other@example.com"), medium.NewSMS("972 12345678")}},
		{Id: 3, Mediums: []medium.Medium{medium.NewSMS("972 87654321")}},
	}
}

func setupRouter(q *queue.Queue) *http.ServeMux {
	mux := http.NewServeMux()

	// Update user notification preferences (which channels to use and frequency)
	mux.HandleFunc("POST /users/preferences", func(w http.ResponseWriter, r *http.Request) {
	})

	// Send a notification to users based on their preferences.
	mux.HandleFunc("POST /notifications", func(w http.ResponseWriter, r *http.Request) {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nr := &notification.Request{}
		err = json.Unmarshal(content, nr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg := notification.New(nr.Type, &nr.Content)
		for _, u := range users {
			q.Push(delivery.New(&u, msg))
		}
	})

	// Retrieve the status of sent notifications (success, failure, retry attempts).
	mux.HandleFunc("GET /notifications/status", func(w http.ResponseWriter, r *http.Request) {
		// SELECT * FROM notifications ORDER BY ts
	})

	return mux
}

func main() {
	addr := flag.String("addr", ":8080", "Listen address")
	flag.Parse()

	q := queue.NewQueue()
	go func() {
		for {
			time.Sleep(25 * time.Millisecond)
			d := q.Pop()
			if d != nil {
				q.Notify(d.User, d.Message)
			}
		}
	}()

	srv := &http.Server{
		Addr:              *addr,
		IdleTimeout:       0,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           httplog.LogRequest(setupRouter(q)),
	}

	fmt.Println(color.HiGreenString("Listening:"), *addr)
	log.Fatal(srv.ListenAndServe())
}

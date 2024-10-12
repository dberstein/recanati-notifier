package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dberstein/recanati-notifier/delivery"
	httplog "github.com/dberstein/recanati-notifier/httplog"
	"github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/user"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3" // Import driver (blank import for registration)
)

var users []user.User
var db *sql.DB

func ensureSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS notifications (id INTEGER PRIMARY KEY, type INTEGER NOT NULL DEFAULT 0, content TEXT);
CREATE TABLE IF NOT EXISTS mediums (id INTEGER PRIMARY KEY, notification_id INTEGER NOT NULL, user_id INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS user_medium (notification_id INTEGER NOT NULL, user_id INTEGER NOT NULL, medium_id INTEGER NOT NULL, content TEXT);
CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name string);
INSERT INTO users (id, name) VALUES (1, "first");
INSERT INTO users (id, name) VALUES (2, "second");
INSERT INTO users (id, name) VALUES (3, "third");
	`); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func NewDb(dsn string) *sql.DB {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	err = ensureSchema(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func init() {
	users = []user.User{
		{Id: 1, Mediums: []medium.Medium{medium.NewEmail("example@example.com")}},
		{Id: 2, Mediums: []medium.Medium{medium.NewEmail("other@example.com"), medium.NewSMS("972 12345678")}},
		{Id: 3, Mediums: []medium.Medium{medium.NewSMS("972 87654321")}},
	}
}

type ListItem struct {
	Id      int    `json:"id"`
	Typ     int    `json:"type"`
	Content string `json:"content"`
}

func setupRouter(dsn string, ch chan *delivery.Delivery) *http.ServeMux {
	db = NewDb(dsn)
	mux := http.NewServeMux()

	// Update user notification preferences (which channels to use and frequency)
	mux.HandleFunc("POST /users/preferences", func(w http.ResponseWriter, r *http.Request) {
	})

	// Send a notification to users based on their preferences.
	stmtInsertNotification, err := db.Prepare(`INSERT INTO notifications (type, content) VALUES (?, ?)`)
	if err != nil {
		panic(err)
	}
	mux.HandleFunc("POST /notifications", func(w http.ResponseWriter, r *http.Request) {
		nr := &notification.Request{}
		err := json.NewDecoder(r.Body).Decode(nr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if nr.Content == "" {
			http.Error(w, "missing content", http.StatusBadRequest)
			return
		}

		fmt.Println("*", nr.String())
		_, err = stmtInsertNotification.Exec(nr.Type, nr.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send notification message to each user...
		msg := notification.New(nr.Type, &nr.Content)
		for _, usr := range users {
			go func(u *user.User, m *notification.Message) {
				ch <- delivery.New(u, m)
			}(&usr, msg)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})

	stmSelectStatusNotification, err := db.Prepare(`SELECT n.id, n.type, n.content FROM notifications n`)
	if err != nil {
		panic(err)
	}

	// Retrieve the status of sent notifications (success, failure, retry attempts).
	mux.HandleFunc("GET /notifications/status", func(w http.ResponseWriter, r *http.Request) {
		rows, err := stmSelectStatusNotification.Query()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		list := []ListItem{}
		for rows.Next() {
			row := ListItem{}
			err = rows.Scan(&row.Id, &row.Typ, &row.Content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			list = append(list, row)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return mux
}

func main() {
	addr := flag.String("addr", ":8080", "Listen address")
	flag.Parse()

	ch := make(chan *delivery.Delivery)
	go func(c chan *delivery.Delivery) {
		for {
			d := <-c
			d.Notify(c)
		}
	}(ch)

	srv := &http.Server{
		Addr:              *addr,
		IdleTimeout:       0,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           httplog.LogRequest(setupRouter(":memory:", ch)),
	}

	fmt.Println(color.HiGreenString("Listening:"), *addr)
	log.Fatal(srv.ListenAndServe())
}

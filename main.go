package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	httplog "github.com/dberstein/recanati-notifier/httplog"
	"github.com/dberstein/recanati-notifier/notification"

	"github.com/fatih/color"
)

var db *sql.DB

type UserPreferences struct {
	UserID    int64 `json:"user_id"`
	Frequency int   `json:"frequency"`
	Mediums   []struct {
		Type   string `json:"type"`
		Target string `json:"target"`
	} `json:"mediums"`
}

func setupRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Update user notification preferences (which channels to use and frequency)
	mux.HandleFunc("POST /users/preferences", func(w http.ResponseWriter, r *http.Request) {
		usrpref := &UserPreferences{}
		err := json.NewDecoder(r.Body).Decode(usrpref)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if usrpref.UserID == 0 {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}

		// Frequency should be > 0 thus only update transaction if field present...
		if usrpref.Frequency != 0 {
			_, err = tx.Exec("UPDATE users SET frequency = ?", usrpref.Frequency)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Delete all mediums then insert sent mediums...
		_, err = tx.Exec("DELETE FROM mediums WHERE uid = ?", usrpref.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, m := range usrpref.Mediums {
			_, err = tx.Exec("INSERT INTO mediums (uid, type, target) VALUES (?, ?, ?)", usrpref.UserID, m.Type, m.Target)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(usrpref)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Send a notification to users based on their preferences.
	mux.HandleFunc("POST /notifications", func(w http.ResponseWriter, r *http.Request) {
		nr := &notification.Request{}
		err := json.NewDecoder(r.Body).Decode(nr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if nr.Subject == "" {
			http.Error(w, "missing subject", http.StatusBadRequest)
			return
		}

		if nr.Body == "" {
			http.Error(w, "missing body", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := tx.Exec("INSERT INTO notifications (type, subject, body) VALUES (?, ?, ?)", nr.Type, nr.Subject, nr.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tx.Commit()

		_, err = res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})

	// Retrieve the status of sent notifications (success, failure, retry attempts).
	mux.HandleFunc("GET /notifications/status", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
    SELECT d.id,
           n.id AS nid,
           n.type AS ntype,
           n.subject,
           n.body,
           d.type,
           d.uid,
           d.target,
           d.status,
           d.attempt
      FROM delivery d
INNER JOIN notifications n ON n.id = d.nid
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		type ListItem struct {
			Id      int    `json:"id"`
			Nid     int    `json:"nid"`
			Ntype   int    `json:"ntype"`
			Type    string `json:"type"`
			Uid     int    `json:"uid"`
			Target  string `json:"target"`
			Status  bool   `json:"status"`
			Attempt int    `json:"attempt"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}

		list := []ListItem{}
		for rows.Next() {
			row := ListItem{}
			err = rows.Scan(
				&row.Id, &row.Nid, &row.Ntype, &row.Subject, &row.Body, &row.Type, &row.Uid,
				&row.Target, &row.Status, &row.Attempt,
			)
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
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	addr := flag.String("addr", ":8080", "Listen address")
	dsn := flag.String("dsn", ":memory:", "Sqlite database DSN")
	flag.Parse()

	db = NewDb(*dsn)
	mux := setupRouter(db)
	entryPoint := httplog.LogRequest(mux)

	srv := &http.Server{
		Addr:              *addr,
		IdleTimeout:       0,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           entryPoint,
	}

	go deliverInLoop(db, 3)

	fmt.Println(color.HiGreenString("Listening:"), *addr)
	log.Fatal(srv.ListenAndServe())
}

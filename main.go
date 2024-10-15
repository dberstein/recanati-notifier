package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	// "github.com/dberstein/recanati-notifier/delivery"

	httplog "github.com/dberstein/recanati-notifier/httplog"

	// "github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	// "github.com/dberstein/recanati-notifier/user"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3" // Import driver (blank import for registration)
)

var db *sql.DB

func ensureSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id   INTEGER PRIMARY KEY,
	name TEXT
);

CREATE TABLE IF NOT EXISTS mediums (
	id     INTEGER PRIMARY KEY,
	uid    INTEGER NOT NULL,
	type   TEXT NOT NULL,
	target TEXT
);

CREATE TABLE IF NOT EXISTS notifications (
	id      INTEGER PRIMARY KEY,
	ts      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	type    INTEGER NOT NULL DEFAULT 0,
	subject TEXT,
	body    TEXT
);

CREATE TABLE IF NOT EXISTS delivery (
	id      INTEGER PRIMARY KEY,
	ts      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	nid     INTEGER NOT NULL,
	uid     INTEGER NOT NULL,
	type    TEXT,
	target  TEXT,
	status  INTEGER NOT NULL DEFAULT 0,
	attempt INTEGER NOT NULL DEFAULT 0
);

INSERT INTO users (id, name) VALUES (1, "first") ON CONFLICT DO NOTHING;
INSERT INTO mediums (uid, type, target) VALUES (1, "email", "example@example.com");
INSERT INTO users (id, name) VALUES (2, "second") ON CONFLICT DO NOTHING;
INSERT INTO mediums (uid, type, target) VALUES (2, "email", "another@example.com");
INSERT INTO mediums (uid, type, target) VALUES (2, "sms", "0123456789");
INSERT INTO users (id, name) VALUES (3, "third") ON CONFLICT DO NOTHING;
INSERT INTO mediums (uid, type, target) VALUES (3, "sms", "9876543210");
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

type ListItem struct {
	Id      int    `json:"notification_id"`
	Type    int    `json:"type"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Dtype   string `json:"dtype"`
	Uid     int    `json:"uid"`
	Target  string `json:"target"`
	Status  bool   `json:"status"`
}

func insertDeliveries(notitifactionId int64) error {
	_, err := db.Exec(`
		INSERT INTO delivery (nid, uid, type, target)
		     SELECT ?, u.id, m.type, m.target
			   FROM mediums m INNER JOIN users u ON u.id = m.uid;
		`, notitifactionId)
	if err != nil {
		return err
	}

	return nil
}

func setupRouter(dsn string) (*http.ServeMux, *sql.DB) {
	db = NewDb(dsn)
	mux := http.NewServeMux()

	// Update user notification preferences (which channels to use and frequency)
	mux.HandleFunc("POST /users/preferences", func(w http.ResponseWriter, r *http.Request) {
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

		notitifactionId, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Message each user's medium...
		err = insertDeliveries(notitifactionId)
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
	SELECT n.id, n.type, n.subject, n.body, d.type AS dtype, d.uid, d.target, d.status
		FROM delivery d
INNER JOIN notifications n ON n.id = d.nid
		`, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		list := []ListItem{}
		for rows.Next() {
			row := ListItem{}
			err = rows.Scan(&row.Id, &row.Type, &row.Subject, &row.Body, &row.Dtype, &row.Uid, &row.Target, &row.Status)
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

	return mux, db
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	addr := flag.String("addr", ":8080", "Listen address")
	dsn := flag.String("dsn", ":memory:", "Sqlite database DSN")
	flag.Parse()

	mux, db := setupRouter(*dsn)
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

	type Delivery struct {
		Nid, Uid                   int
		Typ, Target, Subject, Body string
	}

	go func() {
		last := 0
		for {
			rows, err := db.Query(`
	SELECT d.rowid, d.nid, d.uid, d.type, d.target, n.subject, n.body
	  FROM delivery d
INNER JOIN notifications n ON n.id = d.nid
	 WHERE d.status = ? ORDER BY n.ts
			`, false)
			if err != nil {
				fmt.Println(err)
				continue
			}

			done := []Delivery{}
			d := Delivery{}
			for rows.Next() {
				err := rows.Scan(&last, &d.Nid, &d.Uid, &d.Typ, &d.Target, &d.Subject, &d.Body)
				if err != nil {
					panic(err)
				}

				// Send notification using relevant notifier...
				var notifier notification.Notifier
				switch d.Typ {
				case "email":
					notifier = &notification.Email{To: d.Target}
				case "sms":
					notifier = &notification.SMS{To: d.Target}
				}

				err = notifier.Notify(d.Subject, d.Body)
				if err != nil {
					log.Println("error:", err.Error())
					continue
				}

				done = append(done, d)
			}
			rows.Close()

			time.Sleep(100 * time.Millisecond)
			for _, d := range done {
				_, err := db.Exec("UPDATE delivery SET status = ? WHERE nid = ? AND uid = ? AND type = ? AND target = ?", true, d.Nid, d.Uid, d.Typ, d.Target)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	fmt.Println(color.HiGreenString("Listening:"), *addr)
	log.Fatal(srv.ListenAndServe())
}

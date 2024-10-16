package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/dberstein/recanati-notifier/notification"
)

type Delivery struct {
	Nid     int    `json:"nid"`
	Uid     int    `json/:"uid"`
	Attempt int    `json:"attempt"`
	Typ     string `json:"type"`
	Target  string `json:"target"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func deliverInLoop(db *sql.DB) {
	last := 0
	for {
		rows, err := db.Query(`
SELECT d.rowid,
   d.nid,
   d.uid,
   d.type,
   d.target,
   n.subject,
   n.body,
   d.attempt
FROM delivery d
INNER JOIN notifications n ON n.id = d.nid
WHERE d.status = ?
AND d.attempt < 3
ORDER BY n.ts
		`, false)
		if err != nil {
			fmt.Println(err)
			continue
		}

		done := []*Delivery{}
		retry := []*Delivery{}
		d := Delivery{}
		for rows.Next() {
			err := rows.Scan(&last, &d.Nid, &d.Uid, &d.Typ, &d.Target, &d.Subject, &d.Body, &d.Attempt)
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

			if notifier != nil {
				err = notifier.Notify(d.Subject, d.Body)
				if err != nil {
					log.Println("error:", err.Error())
					retry = append(retry, &d)
					continue
				}
				done = append(done, &d)
			}
		}
		rows.Close()

		stmtUpdate, err := db.Prepare(`
UPDATE delivery
SET status = ?,
attempt = ?
WHERE nid = ?
AND uid = ?
AND TYPE = ?
AND target = ?
		`)
		if err != nil {
			panic(err)
		}

		time.Sleep(100 * time.Millisecond)
		for _, r := range retry {
			r.Attempt++
			_, err := stmtUpdate.Exec(false, d.Attempt, d.Nid, d.Uid, d.Typ, d.Target)
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(100 * time.Millisecond)
		for _, d := range done {
			d.Attempt++
			_, err := stmtUpdate.Exec(true, d.Attempt, d.Nid, d.Uid, d.Typ, d.Target)
			if err != nil {
				panic(err)
			}
		}
	}
}

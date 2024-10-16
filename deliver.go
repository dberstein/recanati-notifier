package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/dberstein/recanati-notifier/notification"

	"github.com/fatih/color"
)

type Delivery struct {
	Id      int    `json:"id"`
	Ntype   int    `json:"ntype"`
	Type    string `json:"type"`
	Attempt int    `json:"attempt"`
	Target  string `json:"target"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func deliverInLoop(db *sql.DB) {
	maxFailedAttempts := 3

	stmtDeliverSelect, err := db.Prepare(`
	SELECT d.id,
		   n.type AS ntype,
		   d.type,
		   d.attempt,
		   d.target,
		   n.subject,
		   n.body
	FROM delivery d
	INNER JOIN notifications n ON n.id = d.nid
	WHERE d.status = ?
	  AND d.attempt < ?
	ORDER BY n.ts
			`)
	if err != nil {
		panic(err)
	}

	for {
		done := []*Delivery{}
		retry := []*Delivery{}

		rows, err := stmtDeliverSelect.Query(false, maxFailedAttempts)
		if err != nil {
			panic(err)
		}

		d := Delivery{}
		for rows.Next() {
			err := rows.Scan(&d.Id, &d.Ntype, &d.Type, &d.Attempt,
				&d.Target, &d.Subject, &d.Body)
			if err != nil {
				panic(err)
			}

			// Send notification using relevant notifier...
			var notifier notification.Notifier
			switch d.Type {
			case "email":
				notifier = &notification.Email{To: d.Target}
			case "sms":
				notifier = &notification.SMS{To: d.Target}
			}

			if notifier != nil {
				err = notifier.Notify(notification.NotificationType(d.Ntype), d.Subject, d.Body)
				if err != nil {
					log.Println(color.HiRedString("error:"), err.Error())
					retry = append(retry, &d)
					continue
				}
			}
			done = append(done, &d)
		}
		rows.Close()

		time.Sleep(500 * time.Millisecond)

		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		for _, r := range retry {
			r.Attempt++

			_, err = tx.Exec(`
UPDATE delivery
SET status = ?,
	attempt = attempt + 1
WHERE id = ?;
			`, false, d.Id)
			if err != nil {
				panic(err)
			}
		}

		for _, d := range done {
			// d.Attempt++

			_, err = tx.Exec(`
UPDATE delivery
SET status = ?,
	attempt = attempt + 1
WHERE id = ?;
			`, true, d.Id)
			if err != nil {
				panic(err)
			}
		}
		tx.Commit()
	}
}

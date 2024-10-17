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
	Status  bool   `json:"status"`
}

func deliverInLoop(db *sql.DB, maxFailedAttempts int) {
	stmtDeliverSelect, err := db.Prepare(`
    SELECT d.id,
		   d.ntype,
		   d.type,
		   d.attempt,
		   d.target,
		   d.subject,
		   d.body,
		   d.status
      FROM deliveries d
     WHERE d.status = ?
       AND d.attempt < ?`)
	if err != nil {
		panic(err)
	}
	defer stmtDeliverSelect.Close()

	for {
		time.Sleep(500 * time.Millisecond)

		rows, err := stmtDeliverSelect.Query(false, maxFailedAttempts)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}

		// Process item and store as `done` or `retry`...
		d, done, retry := Delivery{}, []int{}, []int{}
		for rows.Next() {
			err := rows.Scan(&d.Id, &d.Ntype, &d.Type, &d.Attempt, &d.Target, &d.Subject, &d.Body, &d.Status)
			if err != nil {
				log.Println("ERROR", err)
				continue
			}

			// Send notification using relevant notifier if any...
			if d.Status == true {
				continue
			}
			if notifier := notifierFactory(d.Type, d.Target); notifier != nil {
				err = notifier.Notify(notification.NotificationType(d.Ntype), d.Subject, d.Body)
				if err == nil {
					done = append(done, d.Id)
				} else {
					retry = append(retry, d.Id)
					log.Println(color.HiRedString("error:"), err.Error())
				}
			}

		}
		rows.Close()

		markItems(done, true)
		markItems(retry, false)
	}
}

func markItems(items []int, status bool) {
	stmt, err := db.Prepare(`
	UPDATE delivery
	   SET attempt = attempt + 1,
		   status = ?
	 WHERE id = ?
	   AND status = false;`)
	if err != nil {
		panic(err)
	}

	for _, id := range items {
		_, err := stmt.Exec(status, id)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}
	}
}

func notifierFactory(typ string, target string) notification.Notifier {
	var notifier notification.Notifier
	switch typ {
	case "email":
		notifier = &notification.Email{To: target}
	case "sms":
		notifier = &notification.SMS{To: target}
	}
	return notifier
}

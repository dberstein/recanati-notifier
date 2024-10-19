package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/notification/notifier"

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

type DeliveryStatus struct {
	Id     int  `json:"id"`
	Status bool `json:"status"`
}

func deliverInLoop(db *sql.DB, maxFailedAttempts int) {
	stmt, err := db.Prepare(`
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
	defer stmt.Close()

	for {
		time.Sleep(500 * time.Millisecond)

		rows, err := stmt.Query(false, maxFailedAttempts)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}

		// Process item and store its id and status...
		dones := []*DeliveryStatus{}
		d := Delivery{}
		for rows.Next() {
			err := rows.Scan(&d.Id, &d.Ntype, &d.Type, &d.Attempt, &d.Target, &d.Subject, &d.Body, &d.Status)
			if err != nil {
				log.Println("ERROR", err)
				continue
			}

			// Send notification using relevant notifier if any...
			if n := notifier.Factory(d.Type, d.Target); n != nil {
				err = n.Notify(notification.NotificationType(d.Ntype), d.Subject, d.Body)
				done := DeliveryStatus{Id: d.Id, Status: err == nil}
				if err != nil {
					log.Println(color.HiRedString("error:"), err.Error())
				}
				dones = append(dones, &done)
			}

		}
		err = rows.Close()
		if err != nil {
			log.Println("ERROR", err)
		}

		go func() {
			err := markItems(dones)
			if err != nil {
				log.Println("ERROR", err)
			}
		}()
	}
}

func markItems(items []*DeliveryStatus) error {
	stmt, err := db.Prepare(`
	UPDATE delivery
	   SET attempt = attempt + 1,
		   status = ?
	 WHERE id = ?
	   AND status = false;`)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err := stmt.Exec(item.Status, item.Id)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}
	}
	return nil
}

package deliver

import (
	"database/sql"
	"log"
	"math/rand/v2"
	"time"

	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/notification/notifier"

	"github.com/fatih/color"
)

func Immediate(db *sql.DB, maxFailedAttempts int) {
	time.Sleep(time.Duration(250+rand.IntN(1250)) * time.Millisecond)

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
INNER JOIN users u ON u.id = d.uid
	 WHERE d.status = ?
	   AND u.frequency <= ?
       AND d.attempt < ?
	 LIMIT 10`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(false, 1, maxFailedAttempts)
	if err != nil {
		log.Println("ERROR", err)
		return
	}

	// Process item and store its id and delivery status for stats or retry...
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
			ds := DeliveryStatus{Id: d.Id, Status: err == nil}
			if err != nil {
				log.Println(color.HiRedString("error:"), err.Error())
			}
			dones = append(dones, &ds)
		}

	}
	err = rows.Close()
	if err != nil {
		log.Println("ERROR", err)
	}

	go func() {
		err := markItems(db, dones)
		if err != nil {
			log.Println("ERROR", err)
		}
	}()
}

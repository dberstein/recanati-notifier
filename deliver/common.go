package deliver

import (
	"database/sql"
	"log"
	"math/rand/v2"
	"time"
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

func Loop(db *sql.DB, maxFailedAttempts int) {
	for {
		time.Sleep(time.Duration(250+rand.IntN(1250)) * time.Millisecond)

		Immediate(db, maxFailedAttempts)
		Hourly(db, maxFailedAttempts)
		Daily(db, maxFailedAttempts)
	}
}

func markItems(db *sql.DB, items []*DeliveryStatus) error {
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

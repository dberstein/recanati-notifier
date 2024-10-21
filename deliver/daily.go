package deliver

import (
	"database/sql"
	"log"
	"math/rand/v2"
	"time"
)

func Daily(db *sql.DB, maxFailedAttempts int) {
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
	   AND u.frequency = ?
	   AND day < CAST(UNIXEPOCH() / 86400 AS int)
       AND d.attempt < ?
	 LIMIT 10`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(false, 86400, maxFailedAttempts)
	if err != nil {
		log.Println("ERROR", err)
		return
	}

	d := Delivery{}
	for rows.Next() {
		err := rows.Scan(&d.Id, &d.Ntype, &d.Type, &d.Attempt, &d.Target, &d.Subject, &d.Body, &d.Status)
		if err != nil {
			log.Println("ERROR", err)
			continue
		}

		// todo: send aggregation per user
	}

	err = rows.Close()
	if err != nil {
		log.Println("ERROR", err)
	}
}

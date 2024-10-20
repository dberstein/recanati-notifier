package deliver

import (
	"database/sql"
	"time"
)

func Hourly(db *sql.DB, maxFailedAttempts int) {
	for {
		time.Sleep(500 * time.Millisecond)
	}
}

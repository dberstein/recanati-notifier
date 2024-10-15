package notification

import (
	"errors"
	"log"
	"math/rand/v2"
	"time"
)

type Email struct {
	To string
}

func (e *Email) Notify(subject, body string) error {
	log.Println("email", e.To, subject, body)
	time.Sleep(time.Duration(rand.IntN(1000) * int(time.Millisecond)))
	if rand.IntN(100) > 90 {
		return errors.New("error sending email")
	}
	return nil
}

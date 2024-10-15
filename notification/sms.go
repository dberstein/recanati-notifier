package notification

import (
	"errors"
	"log"
	"math/rand/v2"
	"time"
)

type SMS struct {
	To string
}

func (s *SMS) Notify(subject, body string) error {
	log.Println("sms", s.To, subject, body)
	time.Sleep(time.Duration(rand.IntN(1000) * int(time.Millisecond)))
	if rand.IntN(100) > 90 {
		return errors.New("error sending sms")
	}
	return nil
}

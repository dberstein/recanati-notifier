package notification

import (
	"errors"
	"log"
	"math/rand/v2"
	"time"

	"github.com/fatih/color"
)

type Email struct {
	To string
}

func (e *Email) Notify(typ NotificationType, subject, body string) error {
	log.Println(color.GreenString("Email"), typ.String(), e.To, subject, body)
	time.Sleep(time.Duration(rand.IntN(1000) * int(time.Millisecond)))
	if rand.IntN(100) > 90 {
		return errors.New("error sending email")
	}
	return nil
}

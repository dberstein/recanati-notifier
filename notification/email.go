package notification

import (
	"errors"
	"log"
	"math/rand/v2"

	"github.com/fatih/color"
)

type Email struct {
	To string
}

func (e *Email) Notify(typ NotificationType, subject, body string) error {
	log.Println(color.GreenString("Email"), color.HiGreenString(typ.String()), color.YellowString(e.To), subject, body)
	if rand.IntN(100) > 90 {
		return errors.New("error sending email")
	}
	return nil
}

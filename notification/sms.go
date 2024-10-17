package notification

import (
	"errors"
	"log"
	"math/rand/v2"

	"github.com/fatih/color"
)

type SMS struct {
	To string
}

func (s *SMS) Notify(typ NotificationType, subject, body string) error {
	log.Println(color.GreenString("SMS"), color.HiGreenString(typ.String()), color.YellowString(s.To), subject, body)
	if rand.IntN(100) > 90 {
		return errors.New("error sending sms")
	}
	return nil
}

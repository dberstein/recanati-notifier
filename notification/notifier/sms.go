package notifier

import (
	"errors"
	"log"
	"math/rand/v2"
	"time"

	"github.com/dberstein/recanati-notifier/notification"

	"github.com/fatih/color"
)

type SMS struct {
	To string
}

func (s *SMS) Notify(typ notification.NotificationType, subject, body string) error {
	time.Sleep(time.Duration(rand.IntN(1000) * int(time.Millisecond)))
	log.Println(color.GreenString("SMS"), color.HiGreenString(typ.String()), color.YellowString(s.To), subject, body)
	if rand.IntN(100) > 75 {
		return errors.New("error sending sms")
	}
	return nil
}

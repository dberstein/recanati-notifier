package delivery

import (
	"fmt"

	"github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/user"
	"github.com/fatih/color"
)

type Delivery struct {
	User    *user.User
	Message *notification.Message
}

func New(u *user.User, msg *notification.Message) *Delivery {
	return &Delivery{User: u, Message: msg}
}

func (d *Delivery) Notify(ch chan *Delivery) {
	var logMsg string
	for _, m := range d.User.Mediums {
		var colorPrint = color.GreenString
		if err := m.Notify(d.Message); err != nil {
			colorPrint = color.RedString
			logMsg = err.Error()
			m.SetStatus(medium.StatusRetry)

			if m.Retry() {
				go func() {
					ch <- d
				}()
			}
		} else {
			logMsg = "success"
			m.SetStatus(medium.StatusSuccess)
		}
		fmt.Println(colorPrint(logMsg))
	}

}

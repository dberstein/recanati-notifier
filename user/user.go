package user

import (
	"fmt"

	"github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/fatih/color"
)

type User struct {
	Id      int
	Mediums []medium.Medium
}

func (u *User) Notify(n *notification.Notification) {
	for _, m := range u.Mediums {
		err := m.Notify(n)
		if err != nil {
			fmt.Println(color.RedString(err.Error()))
			m.SetStatus(medium.StatusError)
		} else {
			fmt.Println(color.GreenString("success"))
			m.SetStatus(medium.StatusSuccess)
		}
	}
}

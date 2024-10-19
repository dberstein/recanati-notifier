package notifier

import "github.com/dberstein/recanati-notifier/notification"

func Factory(typ string, target string) notification.Notifier {
	var n notification.Notifier
	switch typ {
	case "email":
		n = &Email{To: target}
	case "sms":
		n = &SMS{To: target}
	}
	return n
}

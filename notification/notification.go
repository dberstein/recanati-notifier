package notification

import "fmt"

type Message struct {
	Type    NotificationType
	Subject string
	Body    string
}

func New(nt NotificationType, subject string, body string) *Message {
	return &Message{
		Type:    nt,
		Subject: subject,
		Body:    body,
	}
}

func (m *Message) String() string {
	return fmt.Sprintf("* (%s) %s: %s", m.Type, m.Subject, m.Body)
}

type NotificationType int

const (
	TypeInformational NotificationType = iota
	TypeAlert
	TypeReminder
)

func (nt NotificationType) String() string {
	return [...]string{"Informational", "Alert", "Reminder"}[nt]
}

func (nt NotificationType) EnumIndex() int {
	return int(nt)
}

type Request struct {
	Type    NotificationType `json:"type"`
	Subject string           `json:"subject"`
	Body    string           `json:"body"`
}

func (r *Request) String() string {
	return fmt.Sprintf("(%s) %s: %s", r.Type, r.Subject, r.Body)
}

type Notifier interface {
	Notify(NotificationType, string, string) error
}

type ListItem struct {
	Id      int    `json:"id"`
	Nid     int    `json:"nid"`
	Ntype   int    `json:"ntype"`
	Type    string `json:"type"`
	Uid     int    `json:"uid"`
	Target  string `json:"target"`
	Status  bool   `json:"status"`
	Attempt int    `json:"attempt"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

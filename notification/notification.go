package notification

import "fmt"

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
	Content string           `json:"content"`
}

type Message struct {
	Type    NotificationType
	Content *string
}

func New(nt NotificationType, content *string) *Message {
	return &Message{
		Type:    nt,
		Content: content,
	}
}

func (n *Message) String() string {
	return fmt.Sprintf("* %s:\n%s\n", n.Type, *n.Content)
}

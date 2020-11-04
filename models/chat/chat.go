package chat

/*
	OVERVIEW: model for the chat table. Every struct attribute is a reference to the destination table.
*/

import (
	"database/sql"
	"time"
)

type Chat struct {
	ID          int            `gorm:"column:id;primaryKey"`
	User        string         `gorm:"column:user"`
	Channel     sql.NullInt64  `gorm:"column:channel"`
	PMRecipient sql.NullString `gorm:"column:pm_recipient"`
	Message     string         `gorm:"column:message"`
	MessageType string         `gorm:"column:message_type"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (c *Chat) TableName() string {
	return "chat"
}

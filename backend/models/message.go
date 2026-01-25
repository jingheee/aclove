package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type MessageResponse struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Message) ToResponse() *MessageResponse {
	return &MessageResponse{
		ID:        m.ID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

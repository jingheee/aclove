package handlers

import (
	"net/http"
	"time"

	"io.lazydoge/aclove/database"
	"io.lazydoge/aclove/models"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	db *database.DB
}

func NewMessageHandler(db *database.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

func (h *MessageHandler) Create(c *gin.Context) {
	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	now := time.Now()
	var id int64
	err := h.db.QueryRow(
		"INSERT INTO messages (content, created_at) VALUES ($1, $2) RETURNING id",
		req.Content, now,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建消息失败"})
		return
	}

	message := &models.Message{
		ID:        id,
		Content:   req.Content,
		CreatedAt: now,
	}

	c.JSON(http.StatusCreated, message.ToResponse())
}

func (h *MessageHandler) List(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, content, created_at FROM messages ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询消息失败"})
		return
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		m := &models.Message{}
		if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取消息失败"})
			return
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理消息失败"})
		return
	}

	response := make([]*models.MessageResponse, len(messages))
	for i, m := range messages {
		response[i] = m.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *MessageHandler) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
		"version": "1.0.0",
	})
}

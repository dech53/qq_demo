package model

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)
//聊天室相关结构体
type Client struct {
	ID     int
	Socket *websocket.Conn
	Send   chan *SendMessage
	Reader *kafka.Reader
}
type SendMessage struct {
	Type int             `json:"type"`
	ID   int          `json:"id"`
	ToID int          `json:"to_id"`
	Data json.RawMessage `json:"data"`
	FileName string          `json:"filename"`
	MimeType string          `json:"mimeType"`
}
type Chat struct {
	Clients    map[int]*Client
	Broadcast  chan *SendMessage
	Register   chan *Client
	Unregister chan *Client
}
type ServerMsg struct {
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Content string `json:"content"`
}

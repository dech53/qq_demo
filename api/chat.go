package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	// "qq_demo/dao"
	"qq_demo/model"
	"qq_demo/utils"
	"strconv"
	"time"
)

// 初始化聊天室
var ChatRoom = model.Chat{
	Clients:    make(map[int]*model.Client),
	Broadcast:  make(chan *model.SendMessage),
	Register:   make(chan *model.Client),
	Unregister: make(chan *model.Client),
}
var id = 1

func StartWsChat(c *gin.Context) {
	// username, _ := c.Get("username")
	// user, _ := dao.GetInfoByPattern("username", username.(string))
	// log.Println(user)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.ResponseFail(c, err.Error(), 404)
		return
	}
	defer conn.Close()
	go ClientTask()
	newClient := &model.Client{
		ID:     id,
		Socket: conn,
		Send:   make(chan *model.SendMessage),
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{"localhost:9092"},
			Topic:          strconv.Itoa(id),
			CommitInterval: 1 * time.Second,
			GroupID:        strconv.Itoa(id), //定义消费组名字
			StartOffset:    kafka.FirstOffset,
		}),
	}
	go ClientStatusCheck(newClient)
	go Read(newClient, &ChatRoom)
	go Write(newClient)
	ChatRoom.Register <- newClient
	id++
	select {}
}

// 管理client行为
func ClientTask() {
	for {
		select {
		case client := <-ChatRoom.Register:
			log.Printf("用户%v已建立新的连接\n", client.ID)
			ChatRoom.Clients[client.ID] = client
			replyMsg := &model.ServerMsg{
				Type:    "Server",
				Code:    1,
				Content: "用户" + strconv.Itoa(client.ID) + "已连接至服务器",
			}
			msg, _ := json.Marshal(replyMsg)
			client.Socket.WriteMessage(websocket.TextMessage, msg)
			go utils.ReadKafka(&ctx, client)
		case client := <-ChatRoom.Unregister:
			log.Printf("用户%v断开连接\n", client.ID)
			if _, ok := ChatRoom.Clients[client.ID]; ok {
				replyMsg := &model.ServerMsg{
					Type:    "Server",
					Code:    2,
					Content: "连接已断开",
				}
				msg, _ := json.Marshal(replyMsg)
				client.Socket.WriteMessage(websocket.TextMessage, msg)
				close(client.Send)
				delete(ChatRoom.Clients, client.ID)
				client.Reader.Close()
			}
		case message := <-ChatRoom.Broadcast:
			client1 := ChatRoom.Clients[message.ToID]
			client2 := ChatRoom.Clients[message.ID]
			if string(message.Data) == `"该用户不是你的好友"` {
				utils.WriteKafka(message, ctx, strconv.Itoa(message.ID))
			} else {
				//消息类型非好友申请
				if message.Type != 4 && message.Type != 5 {
					if client1 == nil {
						utils.WriteKafka(message, ctx, strconv.Itoa(client2.ID))
						utils.WriteKafka(message, ctx, strconv.Itoa(message.ToID))
					} else {
						utils.WriteKafka(message, ctx, strconv.Itoa(client2.ID))
						utils.WriteKafka(message, ctx, strconv.Itoa(client1.ID))
					}
				} else {
					if message.Type == 4 {
						if client1 == nil {
							utils.WriteKafka(message, ctx, strconv.Itoa(message.ToID))
						} else {
							utils.WriteKafka(message, ctx, strconv.Itoa(client1.ID))
						}
					} else {
						utils.WriteKafka(message, ctx, strconv.Itoa(message.ID))
					}

				}
			}

		}
	}
}

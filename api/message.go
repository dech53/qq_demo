package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	// "path/filepath"
	"qq_demo/dao"
	"qq_demo/model"
	// "strconv"
	"github.com/gorilla/websocket"
	"github.com/minio/minio-go/v7"
	"time"
)

// 读取消息
func Read(client *model.Client, chatroom *model.Chat) {
	defer func() {
		client.Socket.Close()
		ChatRoom.Unregister <- client
	}()
	for {
		var msg model.SendMessage
		err := client.Socket.ReadJSON(&msg)
		msg.ID = client.ID
		if err != nil {
			log.Println("写入msg失败")
			return
		}
		if dao.VerifyFriends(&msg) {
			switch msg.Type {
			case 1: // 文本消息
				log.Println("文本消息:", msg.ID, msg.ToID, msg.Data)
				chatroom.Broadcast <- &msg
			case 2: // 图片消息
				var base64Data string
				err := json.Unmarshal(msg.Data, &base64Data)
				if err != nil {
					log.Println("解码 RawMessage 失败:", err)
					return
				}
				names := strings.Split(msg.FileName, ".")
				// 解码文件的 base64 数据
				fileData, _ := base64.StdEncoding.DecodeString(base64Data)
				os.MkdirAll("./uploads/", os.ModePerm)
				// 生成唯一文件名，格式为 "ID_ToID_时间戳.扩展名"
				fileName := fmt.Sprintf("%d_%d_%d.%s", msg.ID, msg.ToID, time.Now().Unix(), names[1])
				filePath := "./uploads/" + fileName          // 保存路径
				err = os.WriteFile(filePath, fileData, 0644) // 写入文件
				if err != nil {
					log.Println("创建目录失败")
				}
				log.Println("文件保存成功，路径:", filePath)
				//上传文件到minio
				filePath = "./uploads/" + fileName
				objectName := filepath.Base(filePath)
				log.Println(objectName)
				log.Println(filePath)
				ctx := context.Background()
				parts := strings.Split(objectName, "_")
				uploadInfo, err := dao.MinioClient.FPutObject(ctx, "test", "/message/"+parts[0]+"/"+parts[1]+"/"+objectName, filePath, minio.PutObjectOptions{ContentType: msg.MimeType})
				if err != nil {
					log.Fatalln("上传文件失败:", err)
				}
				log.Printf("文件上传成功: %s, 大小: %d 字节\n", uploadInfo.Key, uploadInfo.Size)
				//生成文件访问链接
				presignedURL, _ := dao.MinioClient.PresignedGetObject(ctx, "test", "/message/"+parts[0]+"/"+parts[1]+"/"+objectName, time.Minute*10, nil)
				log.Println(presignedURL)
				msg.Data = json.RawMessage([]byte(`"` + presignedURL.String() + `"`))
				// 将转义字符还原
				msg.Data = bytes.Replace(msg.Data, []byte("\\u0026"), []byte("&"), -1)
				chatroom.Broadcast <- &msg
			case 3: // 视频消息
				log.Println("视频消息,URL:", msg.Data)
				//判断文件大小选择如何上传
				//生成文件访问链接
			case 4: //添加好友
				if string(msg.Data) == `"好友申请"` {
					// 逻辑
					flag := dao.AddFriend(&msg)
					if !flag {
						chatroom.Broadcast <- &msg
					}
				} else if string(msg.Data) == `"同意"` || string(msg.Data) == `"忽略"` {
					flag := dao.FriendRequestAction(&msg)
					if flag {
						msg.Type = 1
						msg.Data = json.RawMessage([]byte(`"好友申请通过"`))
					}
					err := dao.AddRelation(msg.ID, msg.ToID)
					if err != nil {
						log.Print("添加好友失败")
						continue
					}
					chatroom.Broadcast <- &msg
				}
			case 5:
				err := dao.DeleteFriend(&msg)
				msg.Data = json.RawMessage([]byte(`"好友删除成功"`))
				if err == nil {
					chatroom.Broadcast <- &msg
				}
			default:
				log.Println("未知消息类型:", msg.Type)
			}
		} else {
			msg.Data = json.RawMessage([]byte(`"该用户不是你的好友"`))
			chatroom.Broadcast <- &msg
		}
	}
}

// 写入消息
func Write(client *model.Client) {
	defer func() {
		client.Socket.Close()
	}()
	for {
		select {
		case msg := <-client.Send:
			msgStr, _ := json.Marshal(msg)
			client.Socket.WriteMessage(websocket.TextMessage, msgStr)
		}
	}
}

// 心跳检测
func ClientStatusCheck(client *model.Client) {
	for {
		time.Sleep(time.Second * 30)
		err := client.Socket.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			ChatRoom.Unregister <- client
			return
		}
	}
}

package utils

import (
	"context"
	"encoding/json"
	"log"
	"qq_demo/model"
	"strconv"
	"time"
	"github.com/segmentio/kafka-go"
)

func WriteKafka(msg *model.SendMessage, ctx context.Context, topic string) {
	writer := kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           2 * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()
	msgBytes, _ := json.Marshal(msg)
	for i := 0; i < 3; i++ {
		//批量写入信息
		if err := writer.WriteMessages(
			ctx,
			kafka.Message{
				Key:   []byte(strconv.Itoa(msg.ID)),
				Value: msgBytes,
			},
		); err != nil {
			if err == kafka.LeaderNotAvailable {
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				log.Printf("写kafka失败:%v\n", err)
			}
		} else {
			break
		}
	}
}
func ReadKafka(ctx *context.Context, client *model.Client) {
	for {
		// 读取 Kafka 消息
		msg, err := client.Reader.ReadMessage(*ctx)
		if err != nil {
			log.Printf("读kafka失败:%v\n", err)
			break
		}
		// 将 msg.Value 转换为 SendMessage 结构体
		var sendMsg model.SendMessage
		if err := json.Unmarshal(msg.Value, &sendMsg); err != nil {
			log.Printf("解析消息失败: %v\n", err)
			continue // 如果解析失败，跳过当前消息，继续读取下一条
		}
		// 将解析后的消息发送到 client.Send 通道
		select {
		case client.Send <- &sendMsg:
		// 成功发送到通道
		default:
			// 如果通道阻塞，处理相关逻辑，如记录日志或丢弃消息
			log.Printf("发送消息到通道失败,ID: %s\n", client.ID)
		}
	}
}

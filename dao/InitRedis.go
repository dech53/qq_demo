package dao

//初始化redis
import (
	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func InitRdb() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})
	
}
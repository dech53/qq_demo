package dao

import (
	"context"
	"log"
	"qq_demo/model"
	"strconv"
	"time"
)

var ctx = context.Background()

// 添加好友请求
func AddFriend(msg *model.SendMessage) bool {
	key := strconv.Itoa(msg.ID) + ":" + strconv.Itoa(msg.ToID)
	success, err := Rdb.SetNX(ctx, key, string(msg.Data), 0).Result()
	if err != nil {
		log.Println("Redis SetNX 错误:", err)
		return false
	}
	if !success {
		log.Println("用户" + strconv.Itoa(msg.ID) + "对用户" + strconv.Itoa(msg.ToID) + "好友请求已经存在")
		return true
	}
	return false
}

// 好友申请操作
func FriendRequestAction(msg *model.SendMessage) bool {
	if string(msg.Data) == `"同意"` {
		Rdb.Del(ctx, strconv.Itoa(msg.ToID)+":"+strconv.Itoa(msg.ID))
		return true
	} else {
		Rdb.Del(ctx, strconv.Itoa(msg.ToID)+":"+strconv.Itoa(msg.ID))
		return false
	}
	return false
}

// 添加好友关系
func AddRelation(id, to_id int) error {
	var relation model.Relation
	relation.ID_1 = id
	relation.ID_2 = to_id
	relation.CreatedAt = time.Now()
	return DB.Create(relation).Error
}
func DeleteFriend(msg *model.SendMessage) error {
	// 构建 SQL 语句
	query := `
    DELETE FROM relations
    WHERE (ID_1 = ? AND ID_2 = ?) OR (ID_1 = ? AND ID_2 = ?)
    `
	// 执行 SQL 语句
	result := DB.Exec(query, msg.ID, msg.ToID, msg.ToID, msg.ID)
	// 检查删除结果
	if result.Error != nil {
		log.Println("删除好友失败:", result.Error)
		return result.Error
	}
	return nil
}
func GetFriends(user *model.User, friends *[]model.User) error {
	var relations []model.Relation
	DB.Where("ID_1 = ?", user.ID).Or("ID_2 = ?", user.ID).Find(&relations)
	for _, relation := range relations {
		var friend model.User

		// 如果 user.ID 是 id_1，那么查找 id_2 对应的用户；反之亦然
		if relation.ID_1 == user.ID {
			DB.Where("id = ?", relation.ID_2).First(&friend)
		} else {
			DB.Where("id = ?", relation.ID_1).First(&friend)
		}
		// 将查找到的好友加入 friends 数组
		*friends = append(*friends, friend)
	}
	return nil
}
func VerifyFriends(msg *model.SendMessage) bool {
	var relation model.Relation
	// 查找包含好友关系的记录
	result := DB.Where("ID_1 = ? AND ID_2 = ?", msg.ID, msg.ToID).
		Or("ID_1 = ? AND ID_2 = ?", msg.ToID, msg.ID).
		Find(&relation)
	if result.RowsAffected > 0 {
		return true
	}
	return false
}

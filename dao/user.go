package dao

import (
	"log"
	"qq_demo/model"
)
//验证用户信息是否合理
func VerifyUser(new_user *model.User) (bool, error) {
	var user model.User
	err := DB.Where("Username = ?", new_user.Username).Or("Email = ?", new_user.Email).First(&user).Error
	log.Println(user)
	if user.Username == "" {
		return true, nil
	}
	return false, err
}
//添加新用户
func AddNewUser(new_user *model.User) bool {
	err := DB.Create(new_user).Error
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
//根据特征获取用户信息
func GetInfoByPattern(pattern, value string) (model.User, error) {
	var user model.User
	err := DB.Where(pattern+" = ?", value).First(&user)
	return user, err.Error
}

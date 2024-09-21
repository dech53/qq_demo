package api

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"qq_demo/dao"
	"qq_demo/middleware"
	"qq_demo/model"
	"qq_demo/utils"
	"time"
)

var ctx = context.Background()

// 用户注册
func UserRegist(c *gin.Context) {
	var new_user model.User
	c.ShouldBindJSON(&new_user)
	new_user.CreatedAt = time.Now()
	log.Println(new_user)
	flag, err := dao.VerifyUser(&new_user)
	if err != nil {
		utils.ResponseFail(c, err.Error(), 400)
		return
	}
	new_user.Password = utils.Md5Code(new_user.Password)
	if flag && dao.AddNewUser(&new_user) {
		utils.ResponseSuccess(c, "用户注册成功", 200)
		return
	} else {
		utils.ResponseFail(c, "用户名或邮箱已被使用", 400)
		return
	}

}

// 用户登录
func UserLogin(c *gin.Context) {
	var new_user model.User
	c.ShouldBindJSON(&new_user)
	var user model.User
	var err error
	flag := (new_user.Username == "")
	if flag {
		user, err = dao.GetInfoByPattern("email", new_user.Email)
	} else {
		user, err = dao.GetInfoByPattern("username", new_user.Username)
	}
	if err != nil {
		utils.ResponseFail(c, err.Error(), 404)
		return
	}
	if utils.Md5Code(new_user.Password) != user.Password {
		utils.ResponseFail(c, "密码错误", 401)
		return
	}
	savedToken, _ := dao.Rdb.Get(ctx, "user:"+user.Username+":login:"+utils.GetUserAgent(c)).Result()
	if savedToken == "" {
		// 生成 JWT token
		claim := model.MyClaims{
			Username: new_user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
				Issuer:    "dech53",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		tokenString, err := token.SignedString(middleware.Secret)
		if err != nil {
			utils.ResponseFail(c, err.Error(), 404)
			return
		}
		_, err = dao.Rdb.SetNX(ctx, "user:"+user.Username+":login:"+utils.GetUserAgent(c), tokenString, time.Hour*2).Result()
		if err != nil {
			utils.ResponseFail(c, "服务器错误", 403)
			return
		}
		utils.ResponseSuccess(c, tokenString, 200)
		return
	}
	utils.ResponseSuccess(c, savedToken, 200)
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	//获取好友列表,个人信息之类
	username, _ := c.Get("username")
	var friends []model.User
	user, _ := dao.GetInfoByPattern("username", username.(string))
	log.Print(user)
	dao.GetFriends(&user, &friends)
	personal := model.Personal{
		Owner:   user,
		Friends: friends,
	}
	utils.ResponseSuccess(c, personal, 200)
}

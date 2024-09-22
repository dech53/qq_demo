package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"qq_demo/dao"
	"qq_demo/middleware"
	"qq_demo/model"
	"qq_demo/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
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
func Upload(c *gin.Context) {
	// 获取文件和其他参数（发送者ID、接收者ID）
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("获取文件失败:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取文件"})
		return
	}

	senderID := c.PostForm("sender_id")     // 发送者ID
	receiverID := c.PostForm("receiver_id") // 接收者ID
	fileExt := filepath.Ext(file.Filename)  // 获取文件扩展名

	// 生成唯一的文件名
	fileName := fmt.Sprintf("%s_%s_%d%s", senderID, receiverID, time.Now().Unix(), fileExt)
	filePath := "./uploads/" + fileName // 本地存储路径

	// 保存文件到本地目录
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		log.Println("保存文件失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}
	log.Println("文件已保存到:", filePath)
	// 上传文件到 MinIO
	ctx := context.Background()
	objectName := fmt.Sprintf("/message/%s/%s/%s", senderID, receiverID, fileName)
	uploadInfo, err := dao.MinioClient.FPutObject(ctx, "test", objectName, filePath, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		log.Println("上传到 MinIO 失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件上传失败"})
		return
	}
	log.Printf("文件上传成功: %s, 大小: %d 字节\n", uploadInfo.Key, uploadInfo.Size)
	// 生成文件访问链接
	presignedURL, err := dao.MinioClient.PresignedGetObject(ctx, "test", objectName, time.Minute*10, nil)
	if err != nil {
		log.Println("生成文件访问链接失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成链接失败"})
		return
	}
	log.Print(presignedURL)
	utils.ResponseSuccess(c, presignedURL, 200)
}

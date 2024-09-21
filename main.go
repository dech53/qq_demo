package main

import (
	"qq_demo/api"
	"qq_demo/dao"
	"qq_demo/model"
)

func main() {
	dao.InitDB()
	dao.InitRdb()
	dao.Initminio()
	dao.DB.AutoMigrate(&model.User{})
	dao.DB.AutoMigrate(&model.Relation{})
	api.InitRouter()
}

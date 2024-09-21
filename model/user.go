package model

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type User struct {
	ID        int    `gorm:"primaryKey;autoIncrement;unique;not null" json:"id"`
	Username  string `gorm:"size:50;not null" json:"username"`
	Password  string `gorm:"size:60;not null" json:"password"`
	Email     string `gorm:"size:255;unique;not null" json:"email"`
	CreatedAt time.Time
}
type Personal struct {
	Owner   User
	Friends []User
}
type Relation struct {
	ID_1      int
	ID_2      int
	CreatedAt time.Time
}
type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

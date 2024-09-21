package model

// 响应类型结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

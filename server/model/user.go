package model

//定义一个用户的结构体

type User struct {
	//为了序列化和反序列化成功 我们必须保证
	//用户信息的json字串的key和结构体字段对应的tag保持一直
	UserId string	`json:"userId"`
	UserPwd string	`json:"UserPwd"`
	UserName string	`Json:"UserName"`
}
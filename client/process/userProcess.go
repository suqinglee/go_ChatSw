package process
import (
	"fmt"
	"net"
	"go_ChatSw/public"
	"encoding/binary"
	"encoding/json"
	"os"
)

type UserProcess struct {
	//暂时不需要字段
}


func (this *UserProcess) Register(userId string,
	userPwd string,userName string) (err error) {

	//1.连接到服务器
	conn,err := net.Dial("tcp","0.0.0.0:8889")
	if err != nil {
		fmt.Println("net.Dial() error:",err)
		return
	}
	defer conn.Close()

	//2.连接成功，准备通过conn发送消息给服务器
	var mes public.Message
	mes.Type = public.RegisterMesType


	var registerMes public.RegisterMes
	registerMes.User.UserId = userId
	registerMes.User.UserPwd = userPwd
	registerMes.User.UserName = userName

	data,err := json.Marshal(registerMes)
	if err != nil {
		fmt.Println("registerMes json.Marshal() error:",err)
		return
	}

	mes.Data = string(data)

	data,err = json.Marshal(mes)
	if err != nil {
		fmt.Println("mes json.Marshal() error:",err)
		return
	}

	tf := &public.Transfer {
		Conn:conn,
	}

	//发送data给服务器
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("注册发送信息出错:",err)
	}

	mes,err = tf.ReadPkg()
	if err != nil {
		fmt.Println("public.ReadPkg(conn) error:",err)
		return
	}

	var registerResMes public.RegisterResMes
	err = json.Unmarshal([]byte(mes.Data),&registerResMes)
	if registerResMes.Code == 200 {
		fmt.Println("注册成功")
		os.Exit(0)
	} else {
		fmt.Println(registerResMes.Error)
		os.Exit(0)
	}
	return
}



//写一个函数，完成登录
func (this *UserProcess) Login(userId string, userPwd string) (err error) {

	//下一步开始定协议
	// fmt.Println("userId:",userID,"userPWd:",userPwd)
	// return nil

	//1.连接到服务器
	conn,err := net.Dial("tcp","0.0.0.0:8889")
	if err != nil {
		fmt.Println("net.Dial() error:",err)
		return
	}
	defer conn.Close()

	//2.连接成功，准备通过conn发送消息给服务器
	var mes public.Message
	mes.Type = public.LoginMesType

	//3.创建LoginMes结构体
	var loginMes public.LoginMes
	loginMes.UserId = userId
	loginMes.UserPwd = userPwd

	//4.将loginMes序列化
	data,err := json.Marshal(loginMes)
	if err != nil {
		fmt.Println("loginMes json.Marshal() error:",err)
		return
	}

	//5.把data赋给mes.Data字段
	mes.Data = string(data) //data是个[]byte切片，转成string

	//6.将mes进行序列化
	data,err = json.Marshal(mes)
	if err != nil {
		fmt.Println("mes json.Marshal() error:",err)
		return
	}

	//注意*:conn.Write() 只能发送[]byte切片

	//7.data是个[]byte切片，这时候data就是我们要发送的消息
	//7.1 先把data的长度发送给服务器
	//先获取到 data的长度 -> 转成一个表示长度的byte切片
	var pkgLen uint32
	pkgLen = uint32(len(data))
	var buf []byte = make([]byte,4,4)
	binary.BigEndian.PutUint32(buf,pkgLen)
	//发送长度
	n,err := conn.Write(buf) //n是发送了多少字节数据
	if n != 4 || err != nil {
		fmt.Println("conn.Write(buf) error:",err)
		return
	}

	// fmt.Println("客户端发送消息的长度成功:",len(data),"内容:",string(data))
	
	//发送消息本身
	_,err = conn.Write(data) //data即序列化后的Message
	if err != nil {
		fmt.Println("conn.Write(data) error:",err)
		return
	}

	//这里还需要处理服务器端返回的消息

	//创建一个实例
	tf := &public.Transfer {
		Conn:conn,
	}
	//
	mes,err = tf.ReadPkg()
	if err != nil {
		fmt.Println("public.ReadPkg(conn) error:",err)
		return
	}
	
	//将mes.Data反序列化程LoginResMes
	var loginResMes public.LoginResMes
	err = json.Unmarshal([]byte(mes.Data),&loginResMes)
	if loginResMes.Code == 200 {

		//初始化CurUser
		CurUser.Conn = conn
		CurUser.UserId = userId
		CurUser.UserStatus = public.UserOnline

		// fmt.Println("登录成功")

		//现在可以显示一下当前在线用户列表，遍历loginResMes.UsersId
		fmt.Println("当前在线用户列表:")
		for _,v := range loginResMes.UsersId {
			fmt.Println("用户id:\t",v)

			//完成onlineUsers初始化
			user := &public.User {
				UserId:v,
				UserStatus:public.UserOnline,
			}
			onlineUsers[v] = user
		}
		fmt.Print("\n\n")

		//这里我们还要在客户端启动一个协程
		//该协程保持和服务器端的通讯
		//接收服务器的推送并显示在客户端
		go serverProcessMes(conn)

		//1.显示我们登录成功后的菜单
		for {
			ShowMenu()
		}
	} else {
		fmt.Println(loginResMes.Error)
	}
	return
}
package process2
import (
	"fmt"
	"net"
	"go_ChatSw/public"
	"encoding/json"
	"go_ChatSw/server/model"

)

type UserProcess struct {
	Conn net.Conn
	//增加一个字段，表示该conn是那个用户的
	UserId string
}

//这里编写通知所有在线用户的方法
//userId要通知其他的在线用户我上线
func (this *UserProcess) NotifyOthersOnlineUser(userId string) {
	
	//遍历onlineUsers,然后一个一个的发送 NotifyUserStatusMes
	for id,up := range userMgr.onlineUsers {
		//过滤掉自己
		if id == userId {
			continue
		}
		//开始通知 单独的写一个方法
		up.NotifyMeOnline(userId)
	}
}

func (this *UserProcess) NotifyMeOnline(userId string) {
	//组装我们的NotifyUserStatusMes
	var mes public.Message
	mes.Type = public.NotifyUserStatusMesType

	var notifyUserStatusMes public.NotifyUserStatusMes
	notifyUserStatusMes.UserId = userId
	notifyUserStatusMes.Status = public.UserOnline

	//将notifyUserStatusMes序列化
	data,err := json.Marshal(notifyUserStatusMes)
	if err != nil {
		fmt.Println("json.Marshal(notifyUserStatusMes) error:",err)
		return
	}

	mes.Data = string(data)
	data,err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal(mes) error:",err)
		return
	}

	//创建Transfer发送
	tf := &public.Transfer {
		Conn:this.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("NotifyMeOnline error:",err)
		return
	}
}



func (this *UserProcess) ServerProcessRegister(mes *public.Message) (err error) {

	var registerMes public.RegisterMes
	err = json.Unmarshal([]byte(mes.Data),&registerMes)
	if err != nil {
		fmt.Println("json.Unmarshal([]byte(mes.Data),&registerMes) error:",err)
		return
	}

	var resMes public.Message
	resMes.Type = public.RegisterResMesType

	var registerResMes public.RegisterResMes

	err = model.MyUserDao.Register(&registerMes.User)

	if err != nil {
		if err  == model.ERROR_USER_EXISTS {
			registerResMes.Code = 505
			registerResMes.Error = model.ERROR_USER_EXISTS.Error()
		} else {
			registerResMes.Code = 506
			registerResMes.Error = "未知错误"
		}
	} else {
		registerResMes.Code = 200
	}

	data,err := json.Marshal(registerResMes)
	if err != nil {
		fmt.Println("json.Marshal(registerResMes) error:",err)
		return
	}

	resMes.Data = string(data)

	//5对resMes序列化，准备发送
	data,err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal(resMes) error:",err)
		return
	}

	//6发送data 我们将其封装到writePkg函数中
	//因为使用分层模式(mvc)，我们先创建一个Tranfer实例，然后读取
	tf := &public.Transfer {
		Conn:this.Conn,
	}
	err = tf.WritePkg(data)
	return

}



//编写一个函数serverProcessLogin,专门处理登录
func (this *UserProcess) ServerProcessLogin(mes *public.Message) (err error) {
	//核心代码
	//1.先从mes中取出mes.Data，并直接反序列化成LoginMes
	var loginMes public.LoginMes
	err = json.Unmarshal([]byte(mes.Data),&loginMes)
	if err != nil {
		fmt.Println("json.Unmarshal([]byte(mes.Data),&loginMes) error:",err)
		return
	}

	//1先声明一个resMes
	var resMes public.Message
	resMes.Type = public.LoginResMesType

	//2在声明一个LoginResMes，并完成赋值
	var loginResMes public.LoginResMes

	//我们需要到redis数据库去完成验证
	//1.使用model.MyUserDao到redis取验证
	user,err := model.MyUserDao.Login(loginMes.UserId,loginMes.UserPwd)
	if err != nil {
		if err == model.ERROR_USER_NOTEXISTS {
			loginResMes.Code = 500 //500状态码表示用户不存在
			loginResMes.Error = err.Error()	
		} else if err == model.ERROR_USER_PWD {
			loginResMes.Code = 403 
			loginResMes.Error = err.Error()
		} else {
			loginResMes.Code = 505 
			loginResMes.Error = "服务器内部错误"
		}

	} else {
		loginResMes.Code = 200
		//这里因为用户登录成功，我们就把该登录成功的用户放入到userMgr中
		//将登录成功的userId赋值给this
		this.UserId = loginMes.UserId
		userMgr.AddOnlineUser(this)
		//通知其他的在线用户
		this.NotifyOthersOnlineUser(loginMes.UserId)
		//将当前在线用户的id放入到loginResMes.UsersId
		//便利UserMgr.onlineUSers
		for id,_ := range userMgr.onlineUsers {
			loginResMes.UsersId = append(loginResMes.UsersId,id)
		}
		fmt.Println(user,"登录成功")
	}



	//如果用户的id为100，密码为123456，认为是正确的
	// if loginMes.UserId == "100" && loginMes.UserPwd == "123456" {
	// 	//合法
	// 	loginResMes.Code = 200
	// } else {
	// 	//不合法
	// 	loginResMes.Code = 500 //500状态码表示用户不存在
	// 	loginResMes.Error = "该用户不存在"
	// }

	//3将loginResMes序列化
	data,err := json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("json.Marshal(loginResMes) error:",err)
		return
	}

	//4将data赋值给resMes
	resMes.Data = string(data)

	//5对resMes序列化，准备发送
	data,err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal(resMes) error:",err)
		return
	}

	//6发送data 我们将其封装到writePkg函数中
	//因为使用分层模式(mvc)，我们先创建一个Tranfer实例，然后读取
	tf := &public.Transfer {
		Conn:this.Conn,
	}
	err = tf.WritePkg(data)
	return

}
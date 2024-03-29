package process
import (
	"fmt"
	"os"
	"go_ChatSw/public"
	"net"
	"encoding/json"
)

//显示登录成功后的界面..
func ShowMenu() {
	fmt.Println("----------登录成功----------")
	fmt.Println("       1 显示在线用户列表")
	fmt.Println("       2 发送消息")
	fmt.Println("       3 信息列表")
	fmt.Println("       4 退出系统")
	fmt.Print("请选择(1 - 4):")
	var key string
	var content string

	smsProcess := &SmsProcess{}

	fmt.Scanln(&key)
	switch key {
	case "1":
		// fmt.Println("显示在线用户列表")
		outputOnlineUser()
	case "2":
		fmt.Println("你想对大家说点什么...")
		fmt.Scanf("%s\n",&content)
		smsProcess.SendGroupMes(content)
	case "3":
		fmt.Println("信息列表")
	case "4":
		fmt.Println("退出系统")
		os.Exit(0)
	default:
		fmt.Println("输入错误")
	}
}

//和服务器端保持通讯
func serverProcessMes(conn net.Conn) {
	//创建一个Transfer实例，不停的读取服务器发送的消息
	tf := &public.Transfer {
		Conn : conn,
	}
	for {
		// fmt.Println("客户端正在等待服务器发送的消息")
		mes,err := tf.ReadPkg()
		if err != nil {
			fmt.Println("tf.ReadPkg() error:",err)
			return
		}

		//如果读取到消息，又是下一步处理逻辑
		switch mes.Type {
		case public.NotifyUserStatusMesType:
			//有人上线了
			//1.取出NotifyUserStatusMes
			var notifyUserStatusMes public.NotifyUserStatusMes
			json.Unmarshal([]byte(mes.Data),&notifyUserStatusMes)
			//2.把这个用户的信息，状态保存到客户端的map中
			updateUserStatus(&notifyUserStatusMes)
		case public.SmsMesType:	//有人群发消息了
			outputGroupMes(&mes)
		default:
			fmt.Println("服务器端返回了一个未知消息")
		}
		// fmt.Println(mes)
	}
}
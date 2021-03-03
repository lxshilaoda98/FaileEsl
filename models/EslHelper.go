package models

import (
	"bytes"
	"fmt"
	. "github.com/0x19/goesl"
	"github.com/fsnotify/fsnotify"
	//db "github.com/n1n1n1_owner/FaileEsl/database"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type EslConfig struct {
	fshost   string
	fsport   uint
	password string
	timeout  int
	allowIP  []string
}

type SipModel struct {
	ip        string
	userAgent string
	contact   string
	count     int
}

//连接到FS，并监听数据
func ConnectionEsl() (config *viper.Viper) {

	config = viper.New()
	config.AddConfigPath("./")
	config.SetConfigName("config")
	config.SetConfigType("json")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := config.ReadInConfig(); err != nil {
			panic(err)
		}
	})

	//直接反序列化为Struct
	var configjson EslConfig
	if err := config.Unmarshal(&configjson); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("connection User:%s, Port:%d , PWd:%s , Tiemout:%d \n , ip:%v",
		config.GetString("EslConfig.fshost"), config.GetUint("EslConfig.fsport"),
		config.GetString("EslConfig.password"), config.GetInt("EslConfig.timeout"), config.GetStringSlice("EslConfig.allowIP"))
	client, err := NewClient(config.GetString("EslConfig.fshost"), config.GetUint("EslConfig.fsport"),
		config.GetString("EslConfig.password"), config.GetInt("EslConfig.timeout"))
	if err != nil {
		fmt.Println("connect Go Esl Failed Err.>", err)
		return
	} else {
		fmt.Println("Connection Success ")
		go client.Handle()
		client.Send("events json ALL")
		fmt.Println("初始化map集合")
		allowIP := config.GetStringSlice("EslConfig.allowIP")

		countryCapitalMap := make(map[string]SipModel)
		for {
			msg, err := client.ReadMessage()
			if err != nil {
				// If it contains EOF, we really dont care...
				if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
					Error("Error while reading Freeswitch message: %s", err)
				}
				break
			}
			switch msg.Headers["Event-Name"] {
			case "HEARTBEAT":
				fmt.Println("============心跳事件begin")
				//心跳的时候看下集合数据
				fmt.Println("map集合为：", countryCapitalMap)
				fmt.Println("查看集合中是否有白名单数据.")
				for _, v := range allowIP {
					//查看白名单是否存在黑名单集合中，如果存在就删除掉
					_, ok := countryCapitalMap[v]
					if ok {
						fmt.Println("存在的白名单，删除集合数据")
						delete(countryCapitalMap, v)
					}
				}
				fmt.Println("============心跳事件end")
			case "CUSTOM":
				ipName := ""
				if msg.Headers["contact"] != "" {
					fmt.Println(msg.Headers["contact"])
					fmt.Println(strings.Split(msg.Headers["contact"], "@")[0])
					ipName = strings.Split(strings.Split(msg.Headers["contact"], "@")[1], ":")[0]
					fmt.Println(ipName)
				}
				switch msg.Headers["Event-Subclass"] {
				case "sofia::pre_register":
					fmt.Printf("【预注册】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					//GetUserID(msg.Headers["from-user"])
					//AddFw(msg,countryCapitalMap,ipName)
				case "sofia::register_attempt":
					fmt.Printf("【注册尝试】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
				case "sofia::unregister":
					fmt.Printf("【注销账号】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
				case "sofia::register":
					fmt.Printf("【注册成功账号】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					delete(countryCapitalMap, ipName)
				case "sofia::register_failure":
					fmt.Println("账号错误..>", msg)
					fmt.Printf("【账号错误】注册ip.>%v .注册sip账号：%v \n 客户端：%v .类型：%v \n",
						msg.Headers["to-host"], msg.Headers["to-user"], msg.Headers["user-agent"], msg.Headers["registration-type"])
					//d = GetUserID(msg.Headers["from-user"])
					if msg.Headers["network-ip"] != "" {
						ipName = msg.Headers["network-ip"]
						AddFw(msg, countryCapitalMap, ipName)
					}
				default:
					//Info("未知事件..>",msg)
				}
			default:
				Info("Got new message: %s", msg)
			}

		}
	}
	return
}

func AddFw(msg *Message, countryCapitalMap map[string]SipModel, ipName string) {
	fmt.Println("发现本次请求为异常数据====>添加/修改map集合！")
	sip := SipModel{
		ip:        "",
		userAgent: "",
		contact:   "",
		count:     0,
	}
	capital, ok := countryCapitalMap[ipName]
	if ok {
		fmt.Println("集合数量为===>", len(countryCapitalMap))
		fmt.Println(ipName, "的发起了异常请求...")
		sip.count = capital.count + 1
		sip.ip = capital.ip
		sip.userAgent = capital.userAgent
		sip.contact = capital.contact
		countryCapitalMap[ipName] = sip
		fmt.Println(ipName, "的异常请求总数量：", countryCapitalMap[ipName].count)
		if countryCapitalMap[ipName].count >= 5 {
			fmt.Println("限制ip：", countryCapitalMap[ipName].ip, "..>")
			delete(countryCapitalMap, ipName)
			addfw(capital.ip)
			fmt.Println("删除..>", ipName, "..现在集合长度为：", len(countryCapitalMap))
		}
	} else {
		fmt.Println(ipName, ",这个IP不存在！")
		sip.ip = ipName
		sip.contact = msg.Headers["contact"]
		sip.userAgent = msg.Headers["user-agent"]
		countryCapitalMap[ipName] = sip
		fmt.Println("添加到map集合=====>end")
	}
}

//func GetUserID(user string) (Number int) {
//	sql := "select count(*) as Number from sipuser where SIPUser =?"
//	rows := db.SqlDB.QueryRow(sql, user)
//	rows.Scan(&Number)
//	return
//}
func addfw(ip string) {
	sysType := runtime.GOOS
	fmt.Println("当前系统：", sysType)
	if sysType == "linux" {
		// LINUX系统
		e, f, f1 := exec_shell(fmt.Sprintf("fail2ban-client set freeswitch banip %v", ip))
		if e != nil {
			fmt.Println("执行命令错误..>", e)
		} else {
			fmt.Println("成功执行命令.>", f, "...>", f1)
		}
	}
	if sysType == "windows" {
		// windows系统
		f, e := exec_cmd(fmt.Sprintf("netsh advfirewall firewall add rule name =\"des_%v\" remoteip=\"%v\" dir=in action=block", ip, ip))
		if e != nil {
			fmt.Println("执行命令错误..>", e)
		} else {
			fmt.Println("成功执行命令.>", f)
		}
	}

}

func exec_shell(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func exec_cmd(command string) (str string, err error) {
	var whoami []byte

	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	//cmd := exec.Command("/bin/bash", "", s)
	fmt.Println("运行的cmd..>", command)
	cmd := exec.Command("cmd", "/C", command)
	if whoami, err = cmd.Output(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 指定参数后过滤换行符
	ret, err := simplifiedchinese.GBK.NewDecoder().String(string(whoami))
	fmt.Println(ret)
	return ret, err
}

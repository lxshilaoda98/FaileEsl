package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/0x19/goesl"
	"github.com/fsnotify/fsnotify"
	log "github.com/go-fastlog/fastlog"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	helper "github.com/n1n1n1_owner/FaileEsl/bin/helper"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
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

type CallModel struct {
	Event_type      string `json:"event_type"`
	Event_mess      string `json:"event_mess"`
	Event_time      int64  `json:"event_time"`
	CallNumber      string `json:"call_number"`
	CalledNumber    string `json:"called_number"`
	CallHangupCause string `json:"call_hangup_cause"`
	AgentStatus     string `json:"agent_status"`
	AgentStatusMsg  string `json:"agent_status_msg"`
	AgentState      string `json:"agent_state"`
	AgentStateMsg   string `json:"agent_state_msg"`
}

//连接到FS，并监听数据
func ConnectionEsl() (config *viper.Viper) {

	outFile, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.SetFlags(log.Flags() | log.Ldebug) //设置默认的日志显示级别，不设置所有级别的日志都会被输出，并且不显示日志级别（为了和官方log包保持一致）

	// 修改默认的日志输出对象
	log.SetOutput(outFile)

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

	client, err := ConnFs(config) //尝试连接到fs
	if err != nil {
		fmt.Println("======", err)
		return
	} else {
		fmt.Println("初始化map集合")
		log.Info("初始化map集合")
		allowIP := config.GetStringSlice("EslConfig.allowIP")
		log.Info("白名单IP：", allowIP)
		countryCapitalMap := make(map[string]SipModel)
		for {
			msg, err := client.ReadMessage()
			//fmt.Println("msg=================",msg)
			log.Info(msg)
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
				log.Info("map集合为：", countryCapitalMap)
				log.Info("查看集合中是否有白名单数据.")
				for _, v := range allowIP {
					//查看白名单是否存在黑名单集合中，如果存在就删除掉
					_, ok := countryCapitalMap[v]
					if ok {
						fmt.Println("存在的白名单，删除集合数据")
						delete(countryCapitalMap, v)
					}
				}
				log.Info("============心跳事件end")
			case "CUSTOM":
				ipName := ""
				if msg.Headers["contact"] != "" {
					log.Info(msg.Headers["contact"])
					log.Info(strings.Split(msg.Headers["contact"], "@")[0])
					ipName = strings.Split(strings.Split(msg.Headers["contact"], "@")[1], ":")[0]
					log.Info(ipName)
				}
				switch msg.Headers["Event-Subclass"] {
				case "sofia::pre_register":
					log.Printf("【预注册】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					//GetUserID(msg.Headers["from-user"])
					//AddFw(msg,countryCapitalMap,ipName)
					if msg.Headers["user-agent"] == "unknown" || msg.Headers["user-agent"] == "" {
						AddFw(msg, countryCapitalMap, ipName)
					}
				case "sofia::register_attempt":
					fmt.Printf("【注册尝试】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
				case "sofia::unregister":
					fmt.Printf("【注销账号】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					callAgent := msg.Headers["from-user"]
					CallModel := CallModel{}
					CallModel.Event_type = "1401"
					CallModel.Event_mess = "注销sip账号"
					CallModel.Event_time = time.Now().Unix()
					CallModel.CalledNumber = callAgent
					InsertRedisMQ(callAgent, CallModel)

				case "sofia::register":
					fmt.Printf("【注册成功账号】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					delete(countryCapitalMap, ipName)
				case "sofia::register_failure":
					fmt.Printf("【账号错误】注册ip.>%v .注册sip账号：%v \n 客户端：%v .类型：%v \n",
						msg.Headers["to-host"], msg.Headers["to-user"], msg.Headers["user-agent"], msg.Headers["registration-type"])
					//d = GetUserID(msg.Headers["from-user"])
					if msg.Headers["network-ip"] != "" {
						ipName = msg.Headers["network-ip"]
						AddFw(msg, countryCapitalMap, ipName)
					}
					callAgent := msg.Headers["to-user"]
					CallModel := CallModel{}
					CallModel.Event_type = "1402"
					CallModel.Event_mess = "sip账号错误"
					CallModel.Event_time = time.Now().Unix()
					CallModel.CalledNumber = callAgent
					InsertRedisMQ(callAgent, CallModel)
				case "sofia::wrong_call_state":
					ipName = msg.Headers["network_ip"]
					fmt.Println("错误的异常呼叫..>", ipName)
					AddFw(msg, countryCapitalMap, ipName)
				case "callcenter::info":
					fmt.Println("处理callcenter的请求..>", msg)
					ccAction := msg.Headers["CC-Action"]
					switch ccAction {
					case "members-count":
						fmt.Println("每次调用队列计数api 以及调用者进入或离开队列时，都会生成此事件")
						queueGroup := msg.Headers["CC-Queue"]
						queueCount := msg.Headers["CC-Count"]
						bExist, _ := db.ClientRedis.HExists("call_queue_group", queueGroup).Result()
						fmt.Println("是否存在组", queueGroup, "...>Ex.>>", bExist)
						if !bExist {
							datas := map[string]interface{}{
								queueGroup: queueCount,
							}
							//添加
							if err := db.ClientRedis.HMSet("call_queue_group", datas).Err(); err != nil {
								log.Fatal(err)
							}
						} else {
							datas := map[string]interface{}{
								queueGroup: queueCount,
							}
							db.ClientRedis.HMSet("call_queue_group", datas)
						}
					case "agent-offering":
						callAni := msg.Headers["CC-Member-CID-Number"]
						callAgent := msg.Headers["CC-Agent"]
						fmt.Printf("呼叫振铃给：%v ,来电号码：%v \n", callAgent, callAni)
						CallModel := CallModel{}
						CallModel.Event_type = "1301"
						CallModel.Event_mess = "坐席振铃"
						CallModel.Event_time = time.Now().Unix()
						CallModel.CallNumber = callAni
						CallModel.CalledNumber = callAgent
						InsertRedisMQ(callAgent, CallModel)
					case "bridge-agent-fail":
						callHangup := msg.Headers["CC-Hangup-Cause"]
						if callHangup == "ORIGINATOR_CANCEL" {
							fmt.Println("发起人放弃电话了..通知坐席")
							callAni := msg.Headers["CC-Member-CID-Number"]
							callAgent := msg.Headers["CC-Agent"]
							callHangupTime := msg.Headers["CC-Agent-Aborted-Time"]
							TimeInt, err := strconv.Atoi(callHangupTime)
							if err != nil {
								fmt.Println("conv int Err..>", err)
							}
							CallModel := CallModel{}
							CallModel.Event_type = "1302"
							CallModel.Event_mess = "呼叫放弃"
							CallModel.Event_time = int64(TimeInt)
							CallModel.CallNumber = callAni
							CallModel.CalledNumber = callAgent
							CallModel.CallHangupCause = callHangup

							InsertRedisMQ(callAgent, CallModel)

						} else {
							fmt.Println("连接失败原因..>", callHangup)
						}
					case "agent-status-change":
						fmt.Println("坐席状态切换")
						agentStatus := msg.Headers["CC-Agent-Status"]
						callAgent := msg.Headers["CC-Agent"]
						CallModel := CallModel{}
						CallModel.Event_type = "1306"
						CallModel.Event_mess = "坐席状态切换"
						CallModel.Event_time = time.Now().Unix()
						CallModel.AgentStatus = agentStatus
						StatusMSG := helper.ConvertCN(agentStatus)
						CallModel.AgentStatusMsg = StatusMSG

						InsertRedisMQ(callAgent, CallModel)
					case "agent-state-change":
						fmt.Println("坐席在队列中的特定状态")
						agentState := msg.Headers["CC-Agent-State"]
						callAgent := msg.Headers["CC-Agent"]
						CallModel := CallModel{}
						CallModel.Event_type = "1307"
						CallModel.Event_mess = "坐席队列状态切换"
						CallModel.Event_time = time.Now().Unix()
						CallModel.AgentState = agentState
						StateMSG := helper.ConvertCN(agentState)
						CallModel.AgentStateMsg = StateMSG

						InsertRedisMQ(callAgent, CallModel)
					}
				case "verto::client_disconnect":
					fmt.Println("freeswitch 服务断开...发起重新连接！")
					time.Sleep(time.Second * 5)
					ConnectionEsl()
				default:
					log.Infof("未知子事件..>%s", msg)
				}
			case "CHANNEL_ANSWER":
				callAgent := msg.Headers["Caller-Callee-ID-Number"]
				callNumber := msg.Headers["Caller-Caller-ID-Number"]                             // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"]                           // 被叫
				callerAnswerTime, _ := strconv.Atoi(msg.Headers["Caller-Channel-Answered-Time"]) //应答时间
				CallModel := CallModel{}
				CallModel.Event_type = "1303"
				CallModel.Event_mess = "电话接起"
				CallModel.Event_time = int64(callerAnswerTime)
				CallModel.CallNumber = callNumber
				CallModel.CalledNumber = callerNumber

				InsertRedisMQ(callAgent, CallModel)
			case "CHANNEL_DESTROY":
				ha := helper.HaHangupV{}
				eventType := "1304"
				eventMsg := "电话销毁"
				callType := msg.Headers["Caller-Logical-Direction"]
				callAgent := msg.Headers["Caller-Callee-ID-Number"]
				callNumber := msg.Headers["Caller-Caller-ID-Number"]   // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"] // 被叫
				callerHangupTime := time.Now().Unix()                  //拒绝时间
				callHangupCause := msg.Headers["Hangup-Cause"]
				ha = helper.ErrConvertCN(callHangupCause)
				if callType == "inbound" {
					if msg.Headers["Caller-Destination-Number"] == "voicemail" {
						fmt.Println("电话销毁，进入了留言!!")
						eventType = "1305"
						eventMsg = "接入语音信箱"

					}
					callAgent = msg.Headers["Caller-Caller-ID-Number"]
					callNumber = msg.Headers["Caller-Callee-ID-Number"]
					callerNumber = msg.Headers["Caller-Caller-ID-Number"]
				}

				CallModel := CallModel{}
				CallModel.Event_type = eventType
				CallModel.Event_mess = eventMsg
				CallModel.Event_time = callerHangupTime
				CallModel.CallNumber = callNumber
				CallModel.CalledNumber = callerNumber
				CallModel.CallHangupCause = ha.HaHangupCauseCause

				InsertRedisMQ(callAgent, CallModel)

			case "CHANNEL_CREATE":
				if msg.Headers["variable_direction"] == "inbound" && msg.Headers["Caller-Context"] == "public" {
					//本逻辑来禁止异常的呼叫ip，如果发现异常的呼叫就加入到黑名单中
					log.Infof(" A leg Call FreeSwitch Inbound")
					CallerAni := msg.Headers["Caller-ANI"]
					CallNumber := msg.Headers["Caller-Destination-Number"]
					CallNetWork := msg.Headers["Caller-Network-Addr"]
					log.Infof("呼叫者%v , 被叫号码：%v", CallerAni, CallNumber)
					if len(CallNumber) > 12 {
						log.Infof("本次呼叫的号码可能异常..>暂时将ip：%v 加入到黑名单！", CallNetWork)
						AddFw(msg, countryCapitalMap, CallNetWork)
					} else {
						log.Infof("呼叫Call：%v. DesCall:%v. CallerIP : %v", CallerAni, CallNumber, CallNetWork)
					}

				}
			default:
				log.Infof("Got new message: %s", msg)
			}
		}
	}
	return
}

func InsertRedisMQ(callAgent string, CallModel CallModel) {
	insRedisByte, err := json.Marshal(CallModel)
	if err != nil {
		fmt.Println("String Convert Byte Err..>", err)
	}
	res, err := db.ClientRedis.RPush("call_event_msg_list_"+callAgent, string(insRedisByte)).Result()
	if err != nil {
		fmt.Println("RPush Err..>", err)
	} else {
		fmt.Println("[存入消息队列MQ]insert Redis Success! res >", res)
		db.ClientRedis.Expire("call_event_msg_list_"+callAgent, time.Hour*2)
	}
}

func ConnFs(config *viper.Viper) (client *Client, err error) {
	fmt.Printf("connection User:%s, Port:%d , PWd:%s , Tiemout:%d \n , ip:%v \n",
		config.GetString("EslConfig.fshost"), config.GetUint("EslConfig.fsport"),
		config.GetString("EslConfig.password"), config.GetInt("EslConfig.timeout"), config.GetStringSlice("EslConfig.allowIP"))

	client, err = NewClient(config.GetString("EslConfig.fshost"), config.GetUint("EslConfig.fsport"),
		config.GetString("EslConfig.password"), config.GetInt("EslConfig.timeout"))
	if err != nil {
		fmt.Println("connect Go Esl Failed Err.>", err)
		fmt.Println("Sleep 10s, reset Connection ESL ..")
		time.Sleep(time.Second * 10)
		ConnFs(config)
	} else {
		fmt.Println("Connection Success")
		log.Info("Connection Success ")

		go client.Handle()
		err := client.Send("events json ALL")
		if err != nil {
			fmt.Println("监听异常..>", err)
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

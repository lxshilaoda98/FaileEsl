package models

import (
	"encoding/xml"
	"fmt"
	. "github.com/0x19/goesl"
	"github.com/fsnotify/fsnotify"
	log "github.com/go-fastlog/fastlog"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	"github.com/n1n1n1_owner/FaileEsl/bin/helper"
	"github.com/spf13/viper"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	XMLName        xml.Name `xml:"result"`
	Text           string   `xml:",chardata"`
	Interpretation struct {
		Text       string `xml:",chardata"`
		Mode       string `xml:"mode,attr"`
		Grammar    string `xml:"grammar,attr"`
		Confidence string `xml:"confidence,attr"`
		Input      struct {
			Text string `xml:",chardata"`
			Mode string `xml:"mode,attr"`
		} `xml:"input"`
		Instance struct {
			Text   string `xml:",chardata"`
			Verify string `xml:"verify,attr"`
			ID     struct {
				Text string `xml:",chardata"`
			} `xml:"id"`
			Asrid struct {
				Text string `xml:",chardata"`
			} `xml:"asrid"`
			Meaning struct {
				Text string `xml:",chardata"`
			} `xml:"meaning"`
		} `xml:"instance"`
	} `xml:"interpretation"`
}

type EslConfig struct {
	fshost       string
	fsport       uint
	password     string
	timeout      int
	allowIP      []string
	openFireWall bool
}

type SipModel struct {
	ip        string
	userAgent string
	contact   string
	count     int
}

type CallModel struct {
	OtherUUid       string `json:"otheruuid"`
	Calluuid        string `json:"calluuid"`
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
	IsTransfer      string `json:"is_transfer"`
}

type TransferCall struct {
	Istrasfer    int
	CCSipUser    string
	CallerNumber string
	CalleeNumber string
	CallType     string
}

var SocketConn = make(map[string]net.Conn)

func handle(conn net.Conn) {
	defer conn.Close()
	// 针对当前连接做发送和接受操作
	for {
		var buf [128]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Printf("read from conn failed, err:%v\n", err)
			break
		}
		recv := string(buf[:n])
		SocketConn[recv] = conn
		fmt.Println("[socket传输]socket:", recv, "..>存入map集合成功！conn..>", conn)
		conn.Write([]byte("OK"))
	}
}

//连接到FS，并监听数据
func ConnectionEsl() (config *viper.Viper) {
	go CreateSocketServer() //新建一个socket服务端
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
			//log.Info(msg)
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
						AddFw(config, msg, countryCapitalMap, ipName)
					}
				case "sofia::register_attempt":
					fmt.Printf("【注册尝试】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
				case "sofia::unregister":
					fmt.Printf("【注销账号】来自ip.>%v .注册sip账号：%v \n 联系地址：%v 域：%v \n 客户端：%v \n",
						ipName, msg.Headers["from-user"], msg.Headers["contact"], msg.Headers["user_context"], msg.Headers["user-agent"])
					callAgent := msg.Headers["from-user"]
					CallModel := CallModel{}
					CallModel.Event_type = "1501"
					CallModel.Event_mess = "注销sip账号"
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.CalledNumber = callAgent

					InsertRedisMQForSipUser(callAgent, CallModel)
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
						AddFw(config, msg, countryCapitalMap, ipName)
					}
					callAgent := msg.Headers["to-user"]
					CallModel := CallModel{}
					CallModel.Event_type = "1502"
					CallModel.Event_mess = "sip账号错误"
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.CalledNumber = callAgent
					InsertRedisMQForSipUser(callAgent, CallModel)
				case "sofia::wrong_call_state":
					ipName = msg.Headers["network_ip"]
					fmt.Println("错误的异常呼叫..>", ipName)
					AddFw(config, msg, countryCapitalMap, ipName)
				case "callcenter::info":
					//fmt.Println("处理callcenter的请求..>", msg)
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
						callUUid := msg.Headers["CC-Member-UUID"]
						callSessionUUid := msg.Headers["CC-Member-Session-UUID"]
						fmt.Printf("呼叫振铃给：%v ,来电号码：%v \n", callAgent, callAni)
						fmt.Printf("uuid：%v ,sessionId：%v \n", callUUid, callSessionUUid)
						calledNumber := AgentSelectContact(callAgent) //通过坐席的工号查询对应的被叫信息
						CallModel := CallModel{}
						CallModel.Calluuid = callSessionUUid
						CallModel.Event_type = "1301"
						CallModel.Event_mess = "坐席振铃"
						CallModel.Event_time = time.Now().UnixNano() / 1e6
						CallModel.CallNumber = callAni
						CallModel.CalledNumber = calledNumber

						InsertRedisMQForAgent(callAgent, CallModel)

					case "bridge-agent-start":
						CCAgent := msg.Headers["CC-Agent"]
						_, err := db.SqlDB.Query("update call_userstatus set CallStatus='呼入应答',CallType='in',CallAnswerTime=Now() where CCAgent=?", CCAgent)
						if err != nil {
							fmt.Println("呼入应答..Err..>", err)
						}
					case "bridge-agent-end":
						CCAgent := msg.Headers["CC-Agent"]
						fmt.Println("呼入销毁===========================", CCAgent)
						fmt.Println("不允许sql..>")
						_, err := db.SqlDB.Query("update call_userstatus set CallStatus='呼叫销毁',CallType=NULL,CallHangupTime=Now() where CCAgent=? and 1=2", CCAgent)
						if err != nil {
							fmt.Println("呼入销毁..Err..>", err)
						} else {
							err := client.BgApi("callcenter_config agent set status " + CCAgent + " 'On Break'")
							if err != nil {
								fmt.Println("bgapi err..>", err)
							} else {
								//继续更新一下坐席的状态为后处理.
								_, err := db.SqlDB.Query("update call_userstatus set OnBreakKey = 1,OnBreakVal='话后',OnBreakTime=Now() where CCAgent = ? ", CCAgent)
								if err != nil {
									fmt.Println("修改坐席话后异常..>Err.>", err)
								} else {
									//查询是否需要切换成空闲状态
									row := db.SqlDB.QueryRow("select AutoReady from call_userstatus where CCAgent=?", CCAgent)
									if err != nil {
										fmt.Println("查询Auto失败.Err..>", err)
									} else {
										var autoReady string
										row.Scan(&autoReady)
										fmt.Println("自动就绪时长为：", autoReady)
										if autoReady != "" {
											go func() {
												se, _ := strconv.Atoi(autoReady)
												if se > 0 {
													time.Sleep(time.Duration(se) * time.Second)
													fmt.Println(se, "秒后，进入空闲")
													err := client.BgApi("callcenter_config agent set status " + CCAgent + " 'Available'")
													if err != nil {
														fmt.Println("bgapi err..>", err)
													}
												} else {
													fmt.Println("无需改变！")
												}
											}()
										}
									}
								}
							}

						}
					case "bridge-agent-fail":
						callHangup := msg.Headers["CC-Hangup-Cause"]
						CCAgent := msg.Headers["CC-Agent"]
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

							InsertRedisMQForAgent(callAgent, CallModel)

						} else {
							fmt.Println("连接失败原因..>", callHangup)
						}
						_, err := db.SqlDB.Query("update call_userstatus set CallStatus='接入失败',CallType=NULL,CallHangupTime=Now() where CCAgent=?", CCAgent)
						if err != nil {
							fmt.Println("接入失败..Err..>", err)
						}
					case "agent-status-change":

						agentStatus := msg.Headers["CC-Agent-Status"]
						callAgent := msg.Headers["CC-Agent"]
						CallModel := CallModel{}
						CallModel.Event_type = "1303"
						CallModel.Event_mess = "坐席状态切换"
						CallModel.Event_time = time.Now().UnixNano() / 1e6
						CallModel.AgentStatus = agentStatus
						StatusMSG := helper.ConvertCN(agentStatus)
						CallModel.AgentStatusMsg = StatusMSG

						InsertRedisMQForAgent(callAgent, CallModel)

					case "agent-state-change":
						fmt.Println("坐席在队列中的特定状态")
						agentState := msg.Headers["CC-Agent-State"]
						callAgent := msg.Headers["CC-Agent"]
						CallModel := CallModel{}
						CallModel.Event_type = "1304"
						CallModel.Event_mess = "坐席队列状态切换"
						CallModel.Event_time = time.Now().UnixNano() / 1e6
						CallModel.AgentState = agentState
						StateMSG := helper.ConvertCN(agentState)
						CallModel.AgentStateMsg = StateMSG

						InsertRedisMQForAgent(callAgent, CallModel)
					}
				case "verto::client_disconnect":
					fmt.Println("freeswitch 服务断开...发起重新连接！")
					res, err := db.SqlDB.Query("select AgentId from agnet_binding")
					if err != nil {
						fmt.Println("查询绑定失败！err ", err)
					} else {
						var AgentIdList []string
						for res.Next() {
							var AgentId string
							res.Scan(&AgentId)
							AgentIdList = append(AgentIdList, AgentId)
						}
						res.Close()
						logout(AgentIdList)
					}

					time.Sleep(time.Second * 5)
					ConnectionEsl()
				case "lua:MrcpEvent":
					fmt.Println("mrcp事件..>")
					uuid := msg.Headers["UUID"]
					result := msg.Headers["MSG"]
					fmt.Println(uuid)
					fmt.Println("result", result)
					if result != "" {
						k := fmt.Sprintf("cdrNew_%v", uuid)
						var v Result
						err = xml.Unmarshal([]byte(result), &v)
						if err != nil {
							panic(err)
						}
						fmt.Println(v.Interpretation.Input.Text)
						fmt.Println(v.Interpretation.Instance.Meaning)
						val := v.Interpretation.Instance.Meaning.Text
						key := fmt.Sprintf("asr_%v", uuid)

						fmt.Println("k=", k, "..>v=", v.Interpretation.Instance.Meaning)

						//插入单条，用来解析每次的说话数据
						nodeKey := fmt.Sprintf("channelASR_%v", uuid)
						_, err := db.ClientRedis.Set(nodeKey, val, time.Second*60).Result()
						if err != nil {
							fmt.Println("插入单ASR Err..>", err)
						}
						oldKey := db.ClientRedis.Get(key).String()
						if oldKey != "" {
							_, err := db.ClientRedis.Append(key, val).Result()
							if err != nil {
								fmt.Println("插入错误：", err)
							}
						} else {
							_, err := db.ClientRedis.Set(key, val, time.Hour*12).Result()
							if err != nil {
								fmt.Println("插入错误：", err)
							}
						}
					}
				case "lua:MrcpEventForChannel":
					fmt.Println("asr语音识别:", msg)
				default:
					log.Infof("未知子事件..>%s", msg)
				}
			case "DETECTED_SPEECH":
				fmt.Println("语音识别总事件：", msg)
			case "RECV_RTCP_MESSAGE":
				fmt.Println("发送rtcp", msg)
			case "CHANNEL_CREATE":
				CallModel := CallModel{}

				callerANI := msg.Headers["Caller-ANI"]
				callerCallerIDNumber := msg.Headers["Caller-Callee-ID-Number"]
				fmt.Println("呼叫创建..>")
				fmt.Println(fmt.Sprintf("主叫：%v . 被叫：%v ", callerANI, callerCallerIDNumber))

				callDirection := msg.Headers["Call-Direction"]
				proFileName := msg.Headers["variable_sofia_profile_name"]
				answerState := msg.Headers["Answer-State"]
				channelCallUUID := msg.Headers["Channel-Call-UUID"]
				callerUniqueID := msg.Headers["Caller-Unique-ID"] //被叫channel id

				fmt.Println(fmt.Sprintf("重要相关参数：Call-Direction: %v \n proFileName: %v ", callDirection, proFileName))
				fmt.Println(fmt.Sprintf("answerState: %v \nchannelCallUUID: %v ", answerState, channelCallUUID))
				fmt.Println(fmt.Sprintf("callerUniqueID: %v ", callerUniqueID))
				if callDirection == "inbound" {
					//内线呼叫fs
					fmt.Println("呼叫id：", channelCallUUID)
					fmt.Println("通知", callerANI, "..>1401 话机振铃")
					CallModel.Calluuid = channelCallUUID
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.Event_type = "1401"
					CallModel.Event_mess = "话机振铃"
					CallModel.CallNumber = callerANI
					CallModel.CalledNumber = callerCallerIDNumber

					InsertRedisMQForSipUser(callerANI, CallModel)
				} else if callDirection == "outbound" {
					fmt.Println("呼叫id：", callerUniqueID)
					if callerANI == "0000000000" {
						callerANI = msg.Headers["Caller-Caller-ID-Number"]
					}
					fmt.Println("真实主叫:", callerANI)
					fmt.Println("查找是否已经存在的通话，如果存在，就证明可能是转接的电话")
					var ttCall TransferCall
					rows := db.SqlDB.QueryRow("select CCSipUser,CallType from call_userstatus where CallerNumber=? and callStatus='转接操作中' ", callerANI)
					rows.Scan(&ttCall.CCSipUser, &ttCall.CallType)
					fmt.Println("是否存在通话：", ttCall)
					CallModel.Calluuid = channelCallUUID
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.CallNumber = callerANI
					CallModel.CalledNumber = callerCallerIDNumber
					if ttCall.CCSipUser != "" {
						CallModel.Event_type = "1701"
						CallModel.Event_mess = "转接振铃"
						fmt.Println("通知", callerANI, "..>1701 转接振铃")
						InsertRedisMQForSipUser(ttCall.CCSipUser, CallModel)

						CallModel.OtherUUid = callerUniqueID
						CallModel.Event_type = "1706"
						CallModel.Event_mess = "转接呼叫"
						InsertRedisMQForSipUser(callerCallerIDNumber, CallModel)
					} else {
						CallModel.Event_type = "1403"
						CallModel.Event_mess = "被叫振铃"
						fmt.Println("通知", callerANI, "..>1403 被叫振铃")
						InsertRedisMQForSipUser(callerANI, CallModel)
					}
					CallModel.Calluuid = callerUniqueID
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.Event_type = "1401"
					CallModel.Event_mess = "话机振铃"
					CallModel.CallNumber = callerANI
					CallModel.CalledNumber = callerCallerIDNumber
					InsertRedisMQForSipUser(callerCallerIDNumber, CallModel)
				}
			case "CHANNEL_ANSWER":
				fmt.Println("呼叫应答..>")
				callDirection := msg.Headers["Call-Direction"]
				channelCallUUID := msg.Headers["Channel-Call-UUID"]
				callerANI := msg.Headers["Caller-ANI"]
				//callerCalleeIDNumber:=msg.Headers["Caller-Callee-ID-Number"]

				fmt.Println("应答方向：", callDirection)

				//otherUUid := msg.Headers["variable_bridge_uuid"]
				callUUid := msg.Headers["variable_call_uuid"]
				//call := msg.Headers["Caller-Caller-ID-Number"]
				callNumber := msg.Headers["Caller-Caller-ID-Number"]   // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"] // 被叫
				CallModel := CallModel{}

				fmt.Println("应答..>", msg.Headers["Other-Leg-Logical-Direction"])
				fmt.Println("主叫..>", callNumber)
				fmt.Println("被叫..>", callerNumber)
				if callDirection == "outbound" {
					fmt.Println("呼叫id：", channelCallUUID)
					if callerANI == "0000000000" {
						callerANI = msg.Headers["Caller-Caller-ID-Number"]
					}

					CallModel.Calluuid = callUUid
					CallModel.Event_time = time.Now().UnixNano()
					CallModel.CallNumber = callNumber
					CallModel.CalledNumber = callerNumber

					var ttCall TransferCall
					rows := db.SqlDB.QueryRow("select CCSipUser,CallType from call_userstatus where CallerNumber=? and CallStatus='转接中' ", callerANI)
					rows.Scan(&ttCall.CCSipUser, &ttCall.CallType)
					fmt.Println("是否存在转接通话：", ttCall)
					if ttCall.CCSipUser != "" {
						CallModel.Event_type = "1702"
						CallModel.Event_mess = "转接接听"
						fmt.Println("通知", ttCall.CCSipUser, "..>1702 转接接听")
						InsertRedisMQForSipUser(ttCall.CCSipUser, CallModel)
					} else {
						CallModel.Event_type = "1404"
						CallModel.Event_mess = "被叫接听"
						fmt.Println("通知", callerANI, "..>1404 被叫接听")
						InsertRedisMQForSipUser(callerANI, CallModel)
					}

					CallModel.Calluuid = callUUid
					CallModel.Event_time = time.Now().UnixNano()
					CallModel.CallNumber = callNumber
					CallModel.CalledNumber = callerNumber
					CallModel.Event_type = "1402"
					CallModel.Event_mess = "话机接起"
					fmt.Println("通知", callerNumber, "..>1402 话机接起")
					InsertRedisMQForSipUser(callerNumber, CallModel)
				}
			case "CHANNEL_DESTROY":
				//fmt.Println("挂机..Msg..>",msg)
				mqNumber := msg.Headers["Caller-Caller-ID-Number"]
				callDirection := msg.Headers["Call-Direction"]
				callNumber := msg.Headers["Caller-ANI"]                // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"] // 被叫
				callUUid := msg.Headers["Channel-Call-UUID"]
				ha := helper.HaHangupV{}
				callModel := CallModel{}
				eventType := "1405"
				eventMsg := "电话销毁"
				otherUUid := msg.Headers["variable_bridge_uuid"]

				callerHangupTime := time.Now().UnixNano() / 1e6 //拒绝时间

				callModel.Calluuid = msg.Headers["Channel-Call-UUID"]
				callModel.Event_time = callerHangupTime
				callModel.Event_type = eventType
				callModel.Event_mess = eventMsg
				callModel.CallHangupCause = ha.HaHangupCauseCause
				callModel.CallNumber = callNumber
				callModel.CalledNumber = callerNumber
				callModel.OtherUUid = otherUUid

				fmt.Println("挂机方向：", callDirection)

				//destName:=msg.Headers["Caller-Caller-ID-Name"]
				//if callDirection == "outbound" {}
				callModel.Calluuid = callUUid

				//------
				if callDirection == "outbound" {
					if msg.Headers["Caller-Callee-ID-Name"] == "Outbound Call" && msg.Headers["Caller-Caller-ID-Name"] != "Outbound Call" {
						//mqNumber=msg.Headers["Caller-Callee-ID-Number"]
						if msg.Headers["Caller-Callee-ID-Number"] == msg.Headers["Caller-Destination-Number"] {
							mqNumber = msg.Headers["Caller-Callee-ID-Number"]
						} else if msg.Headers["Caller-Callee-ID-Number"] != msg.Headers["Caller-Destination-Number"] {
							mqNumber = msg.Headers["Caller-Destination-Number"]
						} else {
							mqNumber = msg.Headers["Caller-Caller-ID-Number"]
						}
					} else if msg.Headers["Caller-Caller-ID-Name"] == "Outbound Call" && msg.Headers["Caller-Caller-ID-Name"] == "Outbound Call" {
						mqNumber = msg.Headers["Caller-Caller-ID-Number"]
						var Transfer = 0
						rows := db.SqlDB.QueryRow("select count(*) as Transfer from call_userstatus where CallType='in' and CCSipUser = ? and CallStatus='转接通话中'", mqNumber)
						rows.Scan(&Transfer)
						if Transfer > 0 {
							fmt.Println("发现是转接电话,通知发起者-->Caller-ANI")
							fmt.Println(msg.Headers["Caller-Callee-ID-Number"], "确认转接..>")
							var QrCallModel = CallModel{}
							QrCallModel.Event_type = "1704"
							QrCallModel.Event_mess = "确认转接"
							InsertRedisMQForSipUser(msg.Headers["Caller-Callee-ID-Number"], QrCallModel)

							fmt.Println(msg.Headers["Caller-Caller-ID-Number"], "转接销毁..>")
							callModel.Event_type = "1703"
							callModel.Event_mess = "转接销毁"
							mqNumber = msg.Headers["Caller-Caller-ID-Number"]
						} else {
							if msg.Headers["Caller-Callee-ID-Number"] == msg.Headers["Caller-Destination-Number"] {
								mqNumber = msg.Headers["Caller-Callee-ID-Number"]
							} else if msg.Headers["Caller-Callee-ID-Number"] != msg.Headers["Caller-Destination-Number"] {
								mqNumber = msg.Headers["Caller-Destination-Number"]
							} else {
								mqNumber = msg.Headers["Caller-Caller-ID-Number"]
							}
						}
					}
				} else {
					if msg.Headers["Caller-Caller-ID-Name"] == "Outbound Call" {
						mqNumber = msg.Headers["Caller-Callee-ID-Number"]
					} else {
						mqNumber = msg.Headers["Caller-Caller-ID-Number"]
					}
				}

				//------
				callModel.Calluuid = callUUid
				fmt.Println("通知", mqNumber, "..>1405 电话销毁")
				//通过销毁的号码查找是否是转接挂断
				var CallerNumber = ""
				var CalleeNumber = ""
				var CallType = ""
				rows := db.SqlDB.QueryRow("select CallerNumber,CalleeNumber,CallType from call_userstatus where OtherChannelNumber = ? and CallStatus in ('转接通话中','转接中') ", mqNumber)
				rows.Scan(&CallerNumber, &CalleeNumber, &CallType)
				if CallerNumber != "" {
					var QrCallModel = CallModel{}
					QrCallModel.Event_type = "1707"
					QrCallModel.Event_mess = "转接挂断"
					if CallType == "in" {
						InsertRedisMQForSipUser(CalleeNumber, QrCallModel)
					} else {
						InsertRedisMQForSipUser(CallerNumber, QrCallModel)
					}

				}

				InsertRedisMQForSipUser(mqNumber, callModel)
			case "CHANNEL_HOLD":
				callUUid := msg.Headers["Channel-Call-UUID"]
				CallModel := CallModel{}
				CallModel.Calluuid = callUUid
				CallModel.Event_type = "1406"
				CallModel.Event_mess = "保持"
				CallModel.Event_time = time.Now().UnixNano() / 1e6
				CallModel.CallNumber = msg.Headers["Caller-Caller-ID-Number"]
				CallModel.CalledNumber = msg.Headers["Caller-Callee-ID-Number"]
				InsertRedisMQForSipUser(CallModel.CallNumber, CallModel)
				InsertRedisMQForSipUser(CallModel.CalledNumber, CallModel)
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='保持中',HoldTime=Now() where ChannelUUid=?", callUUid)
				if err != nil {
					fmt.Println("修改状态为保持..Err..>", err)
				}
			case "CHANNEL_UNHOLD":
				callUUid := msg.Headers["Channel-Call-UUID"]
				CallModel := CallModel{}
				CallModel.Calluuid = callUUid
				CallModel.Event_type = "1407"
				CallModel.Event_mess = "取消保持"
				CallModel.Event_time = time.Now().UnixNano() / 1e6
				CallModel.CallNumber = msg.Headers["Caller-Caller-ID-Number"]
				CallModel.CalledNumber = msg.Headers["Caller-Callee-ID-Number"]
				InsertRedisMQForSipUser(CallModel.CallNumber, CallModel)
				InsertRedisMQForSipUser(CallModel.CalledNumber, CallModel)
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='通话中' where ChannelUUid=?", callUUid)
				if err != nil {
					fmt.Println("修改状态为保持..Err..>", err)
				}
			default:
				fmt.Println("未知事件：", msg.Headers["Event-Name"])
				log.Infof("Got new message: %s", msg)
			}
		}
	}
	return
}

func CreateSocketServer() {
	fmt.Println("新建socket server!")
	tcpServer, _ := net.ResolveTCPAddr("tcp4", ":4466")
	listener, _ := net.ListenTCP("tcp", tcpServer)
	for {
		//当有新的客户端请求来的时候，拿到与客户端的连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("收到socket 连接!")
		//处理逻辑
		go handle(conn)
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

func AddFw(config *viper.Viper, msg *Message, countryCapitalMap map[string]SipModel, ipName string) {
	fmt.Println("发现本次请求为异常数据====>添加/修改map集合！")
	IsOpen := config.GetBool("EslConfig.openFireWall")
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
			if IsOpen == true {
				fmt.Println("开启了防火墙SBC,将添加ip到防火墙中！", countryCapitalMap[ipName].ip)
				delete(countryCapitalMap, ipName)
				addfw(capital.ip)
			} else {
				fmt.Println("没有开启sbc模式..>直接过滤ip.>", countryCapitalMap[ipName].ip)
				delete(countryCapitalMap, ipName)
			}
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

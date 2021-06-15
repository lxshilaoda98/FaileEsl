package models

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/0x19/goesl"
	"github.com/fsnotify/fsnotify"
	log "github.com/go-fastlog/fastlog"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	"github.com/n1n1n1_owner/FaileEsl/bin/helper"
	"github.com/spf13/viper"
)

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
						callUUid := msg.Headers["CC-Member-UUID"]
						callSessionUUid := msg.Headers["CC-Member-Session-UUID"]
						fmt.Printf("呼叫振铃给：%v ,来电号码：%v \n", callAgent, callAni)
						fmt.Printf("uuid：%v ,sessionId：%v \n", callUUid, callSessionUUid)
						CallModel := CallModel{}
						CallModel.Calluuid = callSessionUUid
						CallModel.Event_type = "1301"
						CallModel.Event_mess = "坐席振铃"
						CallModel.Event_time = time.Now().UnixNano() / 1e6
						CallModel.CallNumber = callAni
						CallModel.CalledNumber = callAgent
						InsertRedisMQForAgent(callAgent, CallModel)
						_, err := db.SqlDB.Query("update call_userstatus set CallType='in',CallerNumber=?,CalleeNumber=?,ChannelUUid=?,CallStatus='呼入响铃',CallRingTime=Now() where CCAgent=?",
							callAni, callAgent, callSessionUUid, callAgent)
						if err != nil {
							fmt.Println("呼入坐席振铃..Err..>", err)
						}
					case "bridge-agent-start":
						CCAgent := msg.Headers["CC-Agent"]
						_, err := db.SqlDB.Query("update call_userstatus set CallStatus='呼入应答',CallAnswerTime=Now() where CCAgent=?", CCAgent)
						if err != nil {
							fmt.Println("呼入应答..Err..>", err)
						}
					case "bridge-agent-end":
						CCAgent := msg.Headers["CC-Agent"]
						fmt.Println("呼入销毁===========================", CCAgent)
						_, err := db.SqlDB.Query("update call_userstatus set CallStatus='呼叫销毁',CallType=NULL,CallHangupTime=Now() where CCAgent=?", CCAgent)
						if err != nil {
							fmt.Println("呼入销毁..Err..>", err)
						} else {
							err := client.BgApi("callcenter_config agent set status " + CCAgent + " 'On Break'")
							if err != nil {
								fmt.Println("bgapi err..>", err)
							} else {
								//继续更新一下坐席的状态为后处理.
								_, err := db.SqlDB.Query("update call_userstatus set OnBreakKey = 1,OnBreakVal='话后' where CCAgent = ? ", CCAgent)
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
						fmt.Println("坐席状态切换")
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
						if agentStatus == "Logged Out" {
							_, err := db.SqlDB.Query("update call_userstatus set CallStatus='注销状态',LoggedOutTime=Now() where CCAgent=?", callAgent)
							if err != nil {
								fmt.Println("修改状态为注销..Err..>", err)
							}
						} else if agentStatus == "Available" {
							_, err := db.SqlDB.Query("update call_userstatus set CallStatus='空闲状态',AvailableTime=Now() where CCAgent=?", callAgent)
							if err != nil {
								fmt.Println("修改状态为空闲..Err..>", err)
							}
						} else if agentStatus == "On Break" {
							_, err := db.SqlDB.Query("update call_userstatus set CallStatus='小休状态',OnBreakTime=Now() where CCAgent=?", callAgent)
							if err != nil {
								fmt.Println("修改状态为小休..Err..>", err)
							}
						}

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
				default:
					log.Infof("未知子事件..>%s", msg)
				}
			case "CHANNEL_ANSWER":
				callUUid := msg.Headers["variable_call_uuid"]
				//callAgent := msg.Headers["Caller-Callee-ID-Number"]
				callNumber := msg.Headers["Caller-Caller-ID-Number"]   // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"] // 被叫
				//callerAnswerTime, _ := strconv.Atoi(msg.Headers["Caller-Channel-Answered-Time"]) //应答时间
				CallModel := CallModel{}
				CallModel.Calluuid = callUUid
				CallModel.Event_type = "1402"
				CallModel.Event_mess = "话机接起"
				CallModel.Event_time = time.Now().UnixNano() / 1e6 //int64(callerAnswerTime)
				CallModel.CallNumber = callNumber
				CallModel.CalledNumber = callerNumber
				callAgent := SipSelectAgent(msg.Headers["Caller-Callee-ID-Number"])

				if msg.Headers["Other-Leg-Logical-Direction"] == "inbound" {
					callAgent = SipSelectAgent(msg.Headers["Caller-Caller-ID-Name"])
					CallModel.Event_type = "1404"
					CallModel.Event_mess = "被叫接听"

				}
				//if msg.Headers["Call-Direction"] == "outbound" && msg.Headers["Caller-ANI"] == "0000000000" && msg.Headers["Answer-State"] == "answered" {
				//	callAgent = SipSelectAgent(callAgent)
				//}
				if callAgent != "" {
					if CallModel.Event_type == "1402" {
						_, err := db.SqlDB.Query("update call_userstatus set TPAnswerTime=Now() where ChannelUUid=?", callUUid)
						if err != nil {
							fmt.Println("修改话机接起时间..Err..>", err)
						}
					} else if CallModel.Event_type == "1404" {
						_, err := db.SqlDB.Query("update call_userstatus set CalleeAnswerTime=Now(),CallStatus='通话中' where ChannelUUid=?", callUUid)
						if err != nil {
							fmt.Println("修改被叫接听时间..Err..>", err)
						}
					}
					InsertRedisMQForSipUser(callAgent, CallModel)
				}
			case "CHANNEL_DESTROY":
				//fmt.Println("销毁电话..>",msg.Headers["Caller-Callee-ID-Number"])
				//fmt.Println("销毁电话..>",msg)
				ha := helper.HaHangupV{}
				eventType := "1405"
				eventMsg := "电话销毁"
				callType := msg.Headers["Caller-Logical-Direction"]
				callAgent := SipSelectAgent(msg.Headers["Caller-Callee-ID-Number"])
				callNumber := msg.Headers["Caller-Caller-ID-Number"]   // 主叫号码
				callerNumber := msg.Headers["Caller-Callee-ID-Number"] // 被叫
				callerHangupTime := time.Now().UnixNano() / 1e6        //拒绝时间
				ha = helper.ErrConvertCN(msg.Headers["Hangup-Cause"])
				if callType == "inbound" {
					if msg.Headers["Caller-Destination-Number"] == "voicemail" {
						fmt.Println("电话销毁，进入了留言!!")
						eventType = "1601"
						eventMsg = "接入语音信箱"
						callAgent = msg.Headers["Caller-Caller-ID-Number"]
						callNumber = msg.Headers["Caller-Callee-ID-Number"]
						callerNumber = msg.Headers["Caller-Caller-ID-Number"]
					} else {
						//callAgent = SipSelectAgent(msg.Headers["Caller-Caller-ID-Number"])
						if msg.Headers["variable_sofia_profile_name"] == "internal" {
							eventType = "1405"
							eventMsg = "电话销毁" //只需要判断一个挂断即可
							callAgent = SipSelectAgent(msg.Headers["Caller-Caller-ID-Number"])
						}
					}

				} else {
					//通知坐席，话机挂断
					//eventMsg = "电话销毁"
					if msg.Headers["variable_sofia_profile_name"] == "internal" {
						eventType = "1405"
						eventMsg = "电话销毁" //只需要判断一个挂断即可
						if msg.Headers["Caller-Callee-ID-Name"] == "Outbound Call" {
							//话机未摘机
							callAgent = SipSelectAgent(msg.Headers["Caller-Callee-ID-Number"])
						} else {
							callAgent = SipSelectAgent(msg.Headers["Caller-Caller-ID-Number"])
						}
					}

				}
				CallModel := CallModel{}
				CallModel.Calluuid = msg.Headers["Channel-Call-UUID"]
				CallModel.Event_type = eventType
				CallModel.Event_mess = eventMsg
				CallModel.Event_time = callerHangupTime
				CallModel.CallNumber = callNumber
				CallModel.CalledNumber = callerNumber
				CallModel.CallHangupCause = ha.HaHangupCauseCause

				if callAgent != "" {
					_, err := db.SqlDB.Query("update call_userstatus set CallHangupTime=Now(),CallStatus='呼叫销毁',CallType = NULL where ChannelUUid=?", CallModel.Calluuid)
					if err != nil {
						fmt.Println("修改通话销毁时间..Err..>", err)
					} else {
					}
					InsertRedisMQForSipUser(callAgent, CallModel)
				}
			case "CHANNEL_CREATE":
				if msg.Headers["variable_direction"] == "inbound" && msg.Headers["Caller-Context"] == "public" {
					//呼入过来的数据判断
					//本逻辑来禁止异常的呼叫ip，如果发现异常的呼叫就加入到黑名单中
					log.Infof(" A leg Call FreeSwitch Inbound")
					CallerAni := msg.Headers["Caller-ANI"]
					CallNumber := msg.Headers["Caller-Destination-Number"]
					CallNetWork := msg.Headers["Caller-Network-Addr"]
					log.Infof("呼叫者%v , 被叫号码：%v", CallerAni, CallNumber)
					if len(CallNumber) > 12 {
						log.Infof("本次呼叫的号码可能异常..>暂时将ip：%v 加入到黑名单！", CallNetWork)
						AddFw(config, msg, countryCapitalMap, CallNetWork)
					} else {
						log.Infof("呼叫Call：%v. DesCall:%v. CallerIP : %v", CallerAni, CallNumber, CallNetWork)
					}
				}
				if msg.Headers["variable_sofia_profile_name"] == "internal" && msg.Headers["variable_direction"] == "outbound" && msg.Headers["Caller-ANI"] == "0000000000" && msg.Headers["Answer-State"] == "ringing" {
					//振铃话机
					SipPhone := msg.Headers["Caller-Callee-ID-Number"]
					//通过话机找到现在关联的坐席人员信息...
					AgentId := SipSelectAgent(SipPhone)
					CallModel := CallModel{}
					CallModel.Calluuid = msg.Headers["Channel-Call-UUID"]
					CallModel.Event_type = "1401"
					CallModel.Event_mess = "话机振铃"
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.CallNumber = msg.Headers["Caller-ANI"]
					CallModel.CalledNumber = msg.Headers["Caller-Callee-ID-Number"]

					//写入数据库，外呼数据 ,如果直接用话机外呼，不记录数据

					if AgentId != "" {
						InsertRedisMQForSipUser(AgentId, CallModel)

						_, err := db.SqlDB.Query("update call_userstatus set CallerNumber =? ,CCAgent=?,ChannelUUid=?,TPRingTime=Now(),CallStatus='话机振铃',CallType='out' where CCAgent=?", CallModel.CalledNumber,
							AgentId, CallModel.Calluuid, AgentId)
						if err != nil {
							fmt.Println("makeCall修改状态表Err..>", err)
						}
					} else {
						fmt.Println("话机振铃异常，原因找不到坐席相关的信息！")
					}
				} else if msg.Headers["variable_sofia_profile_name"] == "external" && msg.Headers["variable_direction"] == "outbound" {

					SipPhone := msg.Headers["Caller-Caller-ID-Number"]
					//通过话机找到现在关联的坐席人员信息...
					AgentId := SipSelectAgent(SipPhone)
					CallModel := CallModel{}
					CallModel.Calluuid = msg.Headers["Channel-Call-UUID"]
					CallModel.Event_type = "1403"
					CallModel.Event_mess = "被叫振铃"
					CallModel.Event_time = time.Now().UnixNano() / 1e6
					CallModel.CallNumber = msg.Headers["Caller-Caller-ID-Number"]
					CallModel.CalledNumber = msg.Headers["Caller-Callee-ID-Number"]
					if AgentId != "" {
						_, err := db.SqlDB.Query("update call_userstatus set CalleeRingTime=Now(),CallStatus='呼叫中',CalleeNumber=? where ChannelUUid=?", CallModel.CalledNumber, CallModel.Calluuid)
						if err != nil {
							fmt.Println("修改话机接起时间..Err..>", err)
						}
						InsertRedisMQForSipUser(AgentId, CallModel)
					} else {
						fmt.Println("被叫振铃异常，原因找不到坐席相关的信息！")
					}
				}
			case "CHANNEL_HOLD":
				callUUid := msg.Headers["Channel-Call-UUID"]
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='保持中' where ChannelUUid=?", callUUid)
				if err != nil {
					fmt.Println("修改状态为保持..Err..>", err)
				}
			case "CHANNEL_UNHOLD":
				callUUid := msg.Headers["Channel-Call-UUID"]
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='通话中' where ChannelUUid=?", callUUid)
				if err != nil {
					fmt.Println("修改状态为保持..Err..>", err)
				}
			default:
				log.Infof("Got new message: %s", msg)
			}
		}
	}
	return
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

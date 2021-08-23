package models

import (
	"encoding/json"
	"fmt"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	"time"
)

//插入消息到redis消息队列  呼入->
func InsertRedisMQForAgent(callAgent string, CallModel CallModel) {
	if callAgent != "" {
		switch CallModel.Event_type {
		case "1301":
			_, err := db.SqlDB.Query("update call_userstatus set CallType='in',CallerNumber=?,CalleeNumber=?,ChannelUUid=?,CallStatus='呼入响铃',CallRingTime=Now() where CCAgent=?",
				CallModel.CallNumber, callAgent, CallModel.Calluuid, callAgent)
			if err != nil {
				fmt.Println("呼入坐席振铃..Err..>", err)
			}
		case "1303":
			if CallModel.AgentStatus == "Logged Out" {
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='注销状态',LoggedOutTime=Now() where CCAgent=?", callAgent)
				if err != nil {
					fmt.Println("修改状态为注销..Err..>", err)
				}
			} else if CallModel.AgentStatus == "Available" {
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='空闲状态',AvailableTime=Now() where CCAgent=?", callAgent)
				if err != nil {
					fmt.Println("修改状态为空闲..Err..>", err)
				}
			} else if CallModel.AgentStatus == "On Break" {
				_, err := db.SqlDB.Query("update call_userstatus set CallStatus='小休状态',OnBreakTime=Now() where CCAgent=?", callAgent)
				if err != nil {
					fmt.Println("修改状态为小休..Err..>", err)
				}
			}

		}
		fmt.Println("MQ..>CallAgent.>", callAgent)
		insRedisByte, err := json.Marshal(CallModel)
		if err != nil {
			fmt.Println("String Convert Byte Err..>", err)
		}
		var Token string
		MQStr := string(insRedisByte)
		rows := db.SqlDB.QueryRow("select Token from call_userstatus where CCAgent = ?", callAgent)
		rows.Scan(&Token)
		fmt.Println(callAgent, " -> 添加消息队列：", MQStr, ".>token.>", Token)
		if Token != "" {
			res, err := db.ClientRedis.RPush("call_event_msg_list_"+Token, MQStr).Result()
			if err != nil {
				fmt.Println("RPush Err..>", err)
			} else {
				fmt.Println("[存入消息队列MQ]insert Redis Success! res >", res)
				db.ClientRedis.Expire("call_event_msg_list_"+Token, time.Hour*2)
			}
		} else {
			fmt.Println("token 为空，将不存入队列")
		}
	} else {
		fmt.Println("坐席工号为空，将不存入队列")
	}

}

func InsertRedisMQForSipUser(SipUser string, CallModel CallModel) {

	if SipUser != "" {
		fmt.Println("修改数据Type..>", CallModel.Event_type)
		switch CallModel.Event_type {
		case "1401":
			//话机振铃
			sql := "update call_userstatus set CalleeNumber=?,CallerNumber =?,ChannelUUid=?,TPRingTime=Now(),CallStatus='话机振铃',CallType='out' where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, CallModel.CalledNumber, CallModel.CallNumber, CallModel.Calluuid, SipUser)
			if err != nil {
				fmt.Println("修改话机振铃..Err..>", err)
			}
		case "1402":
			sql := "update call_userstatus set TPAnswerTime=Now(),CallStatus='通话中',ChannelUUid=?,CallerNumber=?,CalleeNumber=? where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, CallModel.Calluuid, CallModel.CallNumber, CallModel.CalledNumber, SipUser)
			if err != nil {
				fmt.Println("修改话机接起..Err..>", err)
			}
		case "1404":
			sql := "update call_userstatus set CalleeAnswerTime=Now(),CallStatus='通话中',ChannelUUid=?,CallerNumber=?,CalleeNumber=? where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, CallModel.Calluuid, CallModel.CallNumber, CallModel.CalledNumber, SipUser)
			if err != nil {
				fmt.Println("修改被叫接起..Err..>", err)
			}
		case "1405":
			sql := "update call_userstatus set OnBreakKey = 1,OnBreakVal='话后',OnBreakTime=Now(),CallStatus='小休状态' where CCSipUser = ?"
			_, err := db.SqlDB.Query(sql, SipUser)
			if err != nil {
				fmt.Println("修改电话销毁..Err..>", err)
			}
		case "1701":
			sql := "update call_userstatus set CallStatus='转接中',OtherChannelNumber=? where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, CallModel.CalledNumber, SipUser)
			if err != nil {
				fmt.Println("修改话机转接中..Err..>", err)
			}
		case "1702":
			sql := "update call_userstatus set CallStatus='转接通话中',OtherChannelNumber=? where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, CallModel.CalledNumber, SipUser)
			if err != nil {
				fmt.Println("修改话机转接通话中..Err..>", err)
			}
		case "1703":
			sql := "update call_userstatus set CallStatus='话后状态' where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, SipUser)
			if err != nil {
				fmt.Println("修改话机转接销毁..Err..>", err)
			}
		case "1704":
			sql := "update call_userstatus set CallStatus='话后状态' where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, SipUser)
			if err != nil {
				fmt.Println("修改话机转接取消..Err..>", err)
			}
		case "1707":
			sql := "update call_userstatus set CallStatus='通话中',OtherChannelNumber='' where CCSipUser=?"
			_, err := db.SqlDB.Query(sql, SipUser)
			if err != nil {
				fmt.Println("修改转接挂断..Err..>", err)
			}
		default:
			fmt.Println("No Run SQL..>")
		}
	}
	fmt.Println("MQ..>SipUSer.>", SipUser)
	insRedisByte, err := json.Marshal(CallModel)
	if err != nil {
		fmt.Println("String Convert Byte Err..>", err)
	}
	MQStr := string(insRedisByte)
	var Token string
	rows := db.SqlDB.QueryRow("select Token from call_userstatus where CCSipUser = ?", SipUser)
	rows.Scan(&Token)
	fmt.Println(SipUser, " -> SIP 添加消息队列：", MQStr, ".>token.>", Token)

	capital, ok := SocketConn[Token]
	if ok {
		fmt.Println("[socket传输]..>找到数据,进行传输通讯..>")
		capital.Write([]byte(MQStr))
	} else {
		fmt.Println("[socket传输]..>未找到要传输的socket..忽略")
	}

	if Token != "" {
		res, err := db.ClientRedis.RPush("call_event_msg_list_"+Token, MQStr).Result()
		if err != nil {
			fmt.Println("RPush Err..>", err)
		} else {
			fmt.Println("[存入消息队列MQ]insert Redis Success! res >", res)
			db.ClientRedis.Expire("call_event_msg_list_"+Token, time.Hour*2)
		}
	} else {
		fmt.Println("找不到登录的信息! token 为空，将不存入队列!")
	}

}

func InsertRedisMQForToken(token string, CallModel CallModel) {

	fmt.Println("MQ..>token.>", token)
	insRedisByte, err := json.Marshal(CallModel)
	if err != nil {
		fmt.Println("String Convert Byte Err..>", err)
	}
	MQStr := string(insRedisByte)

	fmt.Println(" -> SIP 添加消息队列：", MQStr, ".>token.>", token)
	if token != "" {
		res, err := db.ClientRedis.RPush("call_event_msg_list_"+token, MQStr).Result()
		if err != nil {
			fmt.Println("RPush Err..>", err)
		} else {
			fmt.Println("[存入消息队列MQ]insert Redis Success! res >", res)
			db.ClientRedis.Expire("call_event_msg_list_"+token, time.Hour*2)
		}
	} else {
		fmt.Println("token 为空，将不存入队列")
	}

}

func GetSipUser(callNumber, calleeNumber string) (SipUser string) {
	row := db.SqlDB.QueryRow("select CCSipUser from call_userstatus where CallerNumber=? and CalleeNumber=?", callNumber, calleeNumber)
	row.Scan(&SipUser)
	return
}

//退出方法，清理redis缓存和db的binding数据
func logout(AgentId []string) {
	//首先清理redis的登录成功缓存
	//清理db关联数据
	for _, v := range AgentId {
		res, err := db.ClientRedis.Del(fmt.Sprintf("call_login_succ_%v", v)).Result()
		if err != nil {
			fmt.Printf("【登录redis】删除redis缓存数据Err..>%v \n", err)
		} else {
			fmt.Println(" 【登录redis】删除redis缓存数据成功!", res)
			//db.SqlDB.QueryRow("delete from agent_binding where AgentId = ?", v)
		}
	}

}

//通过sip账号查找坐席的工号..
func SipSelectAgent(SipPhone string) (AgentId string) {
	fmt.Printf("查询%v数据 \n", SipPhone)
	row := db.SqlDB.QueryRow("select CCAgent from call_userstatus where CCSipUser = ?", SipPhone)
	row.Scan(&AgentId)
	return
}

//通过坐席id查询sip账号信息
func AgentSelectContact(AgentId string) (Contact string) {
	fmt.Printf("查询%v数据 \n", AgentId)
	row := db.SqlDB.QueryRow("select CCSipUser from call_userstatus where CCAgent = ?", AgentId)
	row.Scan(&Contact)
	return
}

func SipSelectTokenForCUUid(uuid string) (Token string) {
	fmt.Printf("查询%v数据 \n", uuid)
	row := db.SqlDB.QueryRow("select Token from call_userstatus where ChannelUUId = ?", uuid)
	row.Scan(&Token)
	return
}

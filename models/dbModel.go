package models

import (
	"encoding/json"
	"fmt"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	"time"
)

//插入消息到redis消息队列  呼入->
func InsertRedisMQForAgent(callAgent string, CallModel CallModel) {
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

}

func InsertRedisMQForSipUser(SipUser string, CallModel CallModel) {
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

func AgentSelectContact(AgentId string) (Contact string) {
	fmt.Printf("查询%v数据 \n", AgentId)
	row := db.SqlDB.QueryRow("select CCAgent from CCSipUser where CCAgent = ?", AgentId)
	row.Scan(&Contact)
	return
}

func SipSelectTokenForCUUid(uuid string) (Token string) {
	fmt.Printf("查询%v数据 \n", uuid)
	row := db.SqlDB.QueryRow("select Token from call_userstatus where ChannelUUId = ?", uuid)
	row.Scan(&Token)
	return
}

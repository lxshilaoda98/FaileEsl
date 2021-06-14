package models

import (
	"encoding/json"
	"fmt"
	db "github.com/n1n1n1_owner/FaileEsl/bin/database"
	"time"
)

//插入消息到redis消息队列
func InsertRedisMQ(callAgent string, CallModel CallModel) {
	insRedisByte, err := json.Marshal(CallModel)
	if err != nil {
		fmt.Println("String Convert Byte Err..>", err)
	}
	MQStr := string(insRedisByte)
	fmt.Println(callAgent, " -> 添加消息队列：", MQStr)
	res, err := db.ClientRedis.RPush("call_event_msg_list_"+callAgent, MQStr).Result()
	if err != nil {
		fmt.Println("RPush Err..>", err)
	} else {
		fmt.Println("[存入消息队列MQ]insert Redis Success! res >", res)
		db.ClientRedis.Expire("call_event_msg_list_"+callAgent, time.Hour*2)
	}
}

//退出方法，清理redis缓存和db的binding数据
func logout(AgentId []string) {
	//首先清理redis的登录成功缓存
	//清理db关联数据
	for _,v := range AgentId {
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

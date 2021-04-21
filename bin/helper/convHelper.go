package helper

type HaHangupV struct {
	HaHangupCauseCode    int
	HaHangupCauseSipCode string
	HaHangupCauseName    string
	HaHangupCauseCause   string
	HaHangupCauseDes     string
}

func ErrConvertCN(key string) (HaHangupV HaHangupV) {
	switch key {
	case "UNSPECIFIED":
		HaHangupV.HaHangupCauseCode = 0
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "UNSPECIFIED"
		HaHangupV.HaHangupCauseCause = "未指定。没有其他适用的原因代码。"
		HaHangupV.HaHangupCauseDes = "当没有其他代码适用时，通常由路由器给出。此原因通常发生在与原因1，原因88和原因100相同类型的情况下。"

	case "UNALLOCATED_NUMBER":
		HaHangupV.HaHangupCauseCode = 1
		HaHangupV.HaHangupCauseSipCode = "404"
		HaHangupV.HaHangupCauseName = "UNALLOCATED_NUMBER"
		HaHangupV.HaHangupCauseCause = "未分配（未分配）编号[Q.850值1]"
		HaHangupV.HaHangupCauseDes = "此原因表明无法联系被叫方，因为尽管被叫方号码采用有效格式，但当前未分配（分配）该号码。"

	case "NO_ROUTE_TRANSIT_NET":
		HaHangupV.HaHangupCauseCode = 2
		HaHangupV.HaHangupCauseSipCode = "404"
		HaHangupV.HaHangupCauseName = "NO_ROUTE_TRANSIT_NET"
		HaHangupV.HaHangupCauseCause = "没有通往指定公交网络的路线（国内使用）[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明发送此原因的设备已接收到无法通过特定转接网络路由呼叫的请求。发送此原因的设备无法识别传输网络，原因是传输网络不存在，或者因为该特定传输网络（虽然存在）不为发送此原因的设备提供服务。"

	case "NO_ROUTE_DESTINATION":
		HaHangupV.HaHangupCauseCode = 3
		HaHangupV.HaHangupCauseSipCode = "404"
		HaHangupV.HaHangupCauseName = "NO_ROUTE_DESTINATION"
		HaHangupV.HaHangupCauseCause = "没有通往目的地的路线[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明，由于路由呼叫所经过的网络无法满足所需的目的地，因此无法接通被叫方。基于网络支持此原因。"

	case "CHANNEL_UNACCEPTABLE":
		HaHangupV.HaHangupCauseCode = 6
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "CHANNEL_UNACCEPTABLE"
		HaHangupV.HaHangupCauseCause = "频道不可接受[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明，发送实体在此呼叫中无法使用最近识别的信道。"

	case "CALL_AWARDED_DELIVERED":
		HaHangupV.HaHangupCauseCode = 7
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "CALL_AWARDED_DELIVERED"
		HaHangupV.HaHangupCauseCause = "呼叫已授予，正在通过已建立的渠道进行传递[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明用户已被授予来话呼叫，并且该来话呼叫正在连接到已经为该用户建立的用于类似呼叫的信道（例如，分组模式x.25虚拟呼叫）。"

	case "NORMAL_CLEARING":
		HaHangupV.HaHangupCauseCode = 16
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "NORMAL_CLEARING"
		HaHangupV.HaHangupCauseCause = "正常通话清除[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因指示呼叫被清除，因为参与呼叫的用户之一已请求清除呼叫。在正常情况下，此原因的来源不是网络。"

	case "USER_BUSY":
		HaHangupV.HaHangupCauseCode = 17
		HaHangupV.HaHangupCauseSipCode = "486"
		HaHangupV.HaHangupCauseName = "USER_BUSY"
		HaHangupV.HaHangupCauseCause = "用户忙[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因用于指示由于已遇到用户繁忙状况，被叫方无法接听另一个电话。该原因值可以由被叫用户或网络生成。在用户确定用户忙的情况下，注意到用户设备与呼叫兼容。"

	case "NO_USER_RESPONSE":
		HaHangupV.HaHangupCauseCode = 18
		HaHangupV.HaHangupCauseSipCode = "408"
		HaHangupV.HaHangupCauseName = "NO_USER_RESPONSE"
		HaHangupV.HaHangupCauseCause = "没有用户响应[Q.850]"
		HaHangupV.HaHangupCauseDes = "当被叫方在分配的指定时间段内未通过警报或连接指示响应呼叫建立消息时，将使用此原因。"

	case "NO_ANSWER":
		HaHangupV.HaHangupCauseCode = 19
		HaHangupV.HaHangupCauseSipCode = "480"
		HaHangupV.HaHangupCauseName = "NO_ANSWER"
		HaHangupV.HaHangupCauseCause = "用户未回答（警告用户）[Q.850]"
		HaHangupV.HaHangupCauseDes = "当被叫方已收到警报但在规定的时间内未以连接指示做出响应时，将使用此原因。注–此原因不一定由Q.931程序产生，但可能由内部网络计时器产生。"

	case "SUBSCRIBER_ABSENT":
		HaHangupV.HaHangupCauseCode = 20
		HaHangupV.HaHangupCauseSipCode = "480"
		HaHangupV.HaHangupCauseName = "SUBSCRIBER_ABSENT"
		HaHangupV.HaHangupCauseCause = "订户缺席[Q.850]"
		HaHangupV.HaHangupCauseDes = "当移动站已注销，无法与移动站建立无线电联系或个人电信用户暂时无法在任何用户网络接口上寻址时，将使用此原因值。在这种情况下，索非亚SIP通常会提高USER_NOT_register。"

	case "CALL_REJECTED":
		HaHangupV.HaHangupCauseCode = 21
		HaHangupV.HaHangupCauseSipCode = "603"
		HaHangupV.HaHangupCauseName = "CALL_REJECTED"
		HaHangupV.HaHangupCauseCause = "通话被拒[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备不希望接受此呼叫，尽管由于发送该原因的设备既不忙也不兼容，它可能已经接受了呼叫。网络也可能生成此原因，表明由于补充服务约束而导致呼叫已清除。诊断字段可能包含有关补充服务和拒绝原因的其他信息。"

	case "NUMBER_CHANGED":
		HaHangupV.HaHangupCauseCode = 22
		HaHangupV.HaHangupCauseSipCode = "410"
		HaHangupV.HaHangupCauseName = "NUMBER_CHANGED"
		HaHangupV.HaHangupCauseCause = "号码已更改[Q.850]"
		HaHangupV.HaHangupCauseDes = "当不再分配由主叫方指示的被叫方号码时，将此原因返回给主叫方。新的被叫方号码可以选择包含在诊断字段中。如果网络不支持此原因，则原因为：1，应使用未分配（未分配）的号码。"

	case "REDIRECTION_TO_NEW_DESTINATION":
		HaHangupV.HaHangupCauseCode = 23
		HaHangupV.HaHangupCauseSipCode = "410"
		HaHangupV.HaHangupCauseName = "REDIRECTION_TO_NEW_DESTINATION"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "通用ISUP协议机制使用此原因，该通用ISUP协议机制可以由决定将呼叫建立为另一个被叫号码的交换机调用。这样的交换可以通过使用该原因值来调用重定向机制，以请求呼叫中涉及的先前交换以将呼叫路由到新号码。"

	case "EXCHANGE_ROUTING_ERROR":
		HaHangupV.HaHangupCauseCode = 25
		HaHangupV.HaHangupCauseSipCode = "483"
		HaHangupV.HaHangupCauseName = "EXCHANGE_ROUTING_ERROR"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "此原因表明无法到达用户指示的目的地，因为中间交换机由于达到执行跳数计数器过程的限制而释放了呼叫。此原因由中间节点生成，该中间节点在减少跳数计数器值时给出结果0。"

	case "DESTINATION_OUT_OF_ORDER":
		HaHangupV.HaHangupCauseCode = 27
		HaHangupV.HaHangupCauseSipCode = "502"
		HaHangupV.HaHangupCauseName = "DESTINATION_OUT_OF_ORDER"
		HaHangupV.HaHangupCauseCause = "目的地故障[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明，由于目标的接口无法正常运行，因此无法访问用户指示的目标。术语“无法正常运行”表示无法将信号消息传递到远程方。例如远程方的物理层或数据链路层故障，或者用户设备脱机。"

	case "INVALID_NUMBER_FORMAT":
		HaHangupV.HaHangupCauseCode = 28
		HaHangupV.HaHangupCauseSipCode = "484"
		HaHangupV.HaHangupCauseName = "INVALID_NUMBER_FORMAT"
		HaHangupV.HaHangupCauseCause = "无效的数字格式（地址不完整）[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明，由于被叫方号码格式无效或不完整，无法接通被叫方。"

	case "FACILITY_REJECTED":
		HaHangupV.HaHangupCauseCode = 29
		HaHangupV.HaHangupCauseSipCode = "501"
		HaHangupV.HaHangupCauseName = "FACILITY_REJECTED"
		HaHangupV.HaHangupCauseCause = "设施被拒绝[Q.850]"
		HaHangupV.HaHangupCauseDes = "当网络无法提供用户请求的补充服务时，将返回此原因。"

	case "RESPONSE_TO_STATUS_ENQUIRY":
		HaHangupV.HaHangupCauseCode = 30
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "RESPONSE_TO_STATUS_ENQUIRY"
		HaHangupV.HaHangupCauseCause = "对状态查询的回复[Q.850]"
		HaHangupV.HaHangupCauseDes = "当生成STATUS消息的原因是先前收到STATUS INQUIRY时，STATUS消息中会包含此原因。"

	case "NORMAL_UNSPECIFIED":
		HaHangupV.HaHangupCauseCode = 31
		HaHangupV.HaHangupCauseSipCode = "480"
		HaHangupV.HaHangupCauseName = "NORMAL_UNSPECIFIED"
		HaHangupV.HaHangupCauseCause = "正常，未指定[Q.850]"
		HaHangupV.HaHangupCauseDes = "仅当正常类别中没有其他原因适用时，才使用此原因报告正常事件。"

	case "NORMAL_CIRCUIT_CONGESTION":
		HaHangupV.HaHangupCauseCode = 34
		HaHangupV.HaHangupCauseSipCode = "503"
		HaHangupV.HaHangupCauseName = "NORMAL_CIRCUIT_CONGESTION"
		HaHangupV.HaHangupCauseCause = "无可用的电路/通道[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明当前没有合适的电路/通道可用于处理呼叫。"

	case "NETWORK_OUT_OF_ORDER":
		HaHangupV.HaHangupCauseCode = 38
		HaHangupV.HaHangupCauseSipCode = "502"
		HaHangupV.HaHangupCauseName = "NETWORK_OUT_OF_ORDER"
		HaHangupV.HaHangupCauseCause = "网络故障[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明网络无法正常运行，并且该状况可能会持续较长的时间，例如，立即重新尝试通话不太可能成功。"

	case "NORMAL_TEMPORARY_FAILURE":
		HaHangupV.HaHangupCauseCode = 41
		HaHangupV.HaHangupCauseSipCode = "503"
		HaHangupV.HaHangupCauseName = "NORMAL_TEMPORARY_FAILURE"
		HaHangupV.HaHangupCauseCause = "暂时故障[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明网络无法正常运行，并且该状况不太可能持续很长时间；例如，用户可能希望几乎立即尝试另一个呼叫尝试。"

	case "SWITCH_CONGESTION":
		HaHangupV.HaHangupCauseCode = 42
		HaHangupV.HaHangupCauseSipCode = "503"
		HaHangupV.HaHangupCauseName = "SWITCH_CONGESTION"
		HaHangupV.HaHangupCauseCause = "交换设备拥塞[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明产生此原因的交换设备正在经历高流量时段。"

	case "ACCESS_INFO_DISCARDED":
		HaHangupV.HaHangupCauseCode = 43
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "ACCESS_INFO_DISCARDED"
		HaHangupV.HaHangupCauseCause = "访问信息被丢弃[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明网络无法按照请求将访问信息传递给远程用户，即，用户对用户的信息，低层兼容性，高层兼容性或诊断中指示的子地址。注意，被丢弃的特定类型的访问信息可选地包括在诊断中。"

	case "REQUESTED_CHAN_UNAVAIL":
		HaHangupV.HaHangupCauseCode = 44
		HaHangupV.HaHangupCauseSipCode = "503"
		HaHangupV.HaHangupCauseName = "REQUESTED_CHAN_UNAVAIL"
		HaHangupV.HaHangupCauseCause = "请求的电路/通道不可用[Q.850]"
		HaHangupV.HaHangupCauseDes = "当接口的另一端无法提供请求实体指示的电路或通道时，将返回此原因。"

	case "PRE_EMPTED":
		HaHangupV.HaHangupCauseCode = 45
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "PRE_EMPTED"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "FACILITY_NOT_SUBSCRIBED":
		HaHangupV.HaHangupCauseCode = 50
		HaHangupV.HaHangupCauseSipCode = "486"
		HaHangupV.HaHangupCauseName = "FACILITY_NOT_SUBSCRIBED"
		HaHangupV.HaHangupCauseCause = "请求的设施未订阅[Q.850	]"
		HaHangupV.HaHangupCauseDes = "此原因表明用户已请求可用的补充服务，但无权使用该用户。"

	case "OUTGOING_CALL_BARRED":
		HaHangupV.HaHangupCauseCode = 52
		HaHangupV.HaHangupCauseSipCode = "403"
		HaHangupV.HaHangupCauseName = "OUTGOING_CALL_BARRED"
		HaHangupV.HaHangupCauseCause = "禁止拨出电话"
		HaHangupV.HaHangupCauseDes = "此原因表明，尽管主叫方是传出CUG呼叫的CUG成员，但不允许该CUG成员进行传出呼叫。"

	case "INCOMING_CALL_BARRED":
		HaHangupV.HaHangupCauseCode = 54
		HaHangupV.HaHangupCauseSipCode = "403"
		HaHangupV.HaHangupCauseName = "INCOMING_CALL_BARRED"
		HaHangupV.HaHangupCauseCause = "禁止来电"
		HaHangupV.HaHangupCauseDes = "此原因表明，尽管被叫方是传入CUG呼叫的CUG成员，但不允许传入呼叫到CUG的该成员。"

	case "BEARERCAPABILITY_NOTAUTH":
		HaHangupV.HaHangupCauseCode = 57
		HaHangupV.HaHangupCauseSipCode = "403"
		HaHangupV.HaHangupCauseName = "BEARERCAPABILITY_NOTAUTH"
		HaHangupV.HaHangupCauseCause = "承载能力未经授权[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明用户已请求由产生此原因的设备实现的承载能力，但用户无权使用。"

	case "BEARERCAPABILITY_NOTAVAIL":
		HaHangupV.HaHangupCauseCode = 58
		HaHangupV.HaHangupCauseSipCode = "503"
		HaHangupV.HaHangupCauseName = "BEARERCAPABILITY_NOTAVAIL"
		HaHangupV.HaHangupCauseCause = "承载能力目前不可用[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明用户已请求承载能力，该能力由产生此原因的设备实现，但目前不可用。"

	case "SERVICE_UNAVAILABLE":
		HaHangupV.HaHangupCauseCode = 63
		HaHangupV.HaHangupCauseSipCode = "486"
		HaHangupV.HaHangupCauseName = "SERVICE_UNAVAILABLE"
		HaHangupV.HaHangupCauseCause = "服务或选件不可用，未指定[Q.850]"
		HaHangupV.HaHangupCauseDes = "仅当服务或选件不可用类别中没有其他原因适用时，才使用此原因来报告服务或选件不可用事件。"

	case "BEARERCAPABILITY_NOTIMPL":
		HaHangupV.HaHangupCauseCode = 65
		HaHangupV.HaHangupCauseSipCode = "488"
		HaHangupV.HaHangupCauseName = "BEARERCAPABILITY_NOTIMPL"
		HaHangupV.HaHangupCauseCause = "承载能力未实现[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明发送此原因的设备不支持请求的承载能力。"

	case "CHAN_NOT_IMPLEMENTED":
		HaHangupV.HaHangupCauseCode = 66
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "CHAN_NOT_IMPLEMENTED"
		HaHangupV.HaHangupCauseCause = "通道类型未实现[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明发送此原因的设备不支持所请求的信道类型"

	case "FACILITY_NOT_IMPLEMENTED":
		HaHangupV.HaHangupCauseCode = 69
		HaHangupV.HaHangupCauseSipCode = "501"
		HaHangupV.HaHangupCauseName = "FACILITY_NOT_IMPLEMENTED"
		HaHangupV.HaHangupCauseCause = "请求的设施未实现[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明发送此原因的设备不支持所请求的补充服务。"

	case "SERVICE_NOT_IMPLEMENTED":
		HaHangupV.HaHangupCauseCode = 79
		HaHangupV.HaHangupCauseSipCode = "501"
		HaHangupV.HaHangupCauseName = "SERVICE_NOT_IMPLEMENTED"
		HaHangupV.HaHangupCauseCause = "未实施服务或选项，未指定[Q.850]	"
		HaHangupV.HaHangupCauseDes = "仅当服务或未实现选项类别中没有其他原因适用时，才使用此原因来报告服务或未实现选项事件。"

	case "INVALID_CALL_REFERENCE":
		HaHangupV.HaHangupCauseCode = 81
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "INVALID_CALL_REFERENCE"
		HaHangupV.HaHangupCauseCause = "无效的呼叫参考值[Q.850]	"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已收到带有呼叫参考的消息，该消息当前在用户网络接口上未使用。"

	case "INCOMPATIBLE_DESTINATION":
		HaHangupV.HaHangupCauseCode = 88
		HaHangupV.HaHangupCauseSipCode = "488"
		HaHangupV.HaHangupCauseName = "INCOMPATIBLE_DESTINATION"
		HaHangupV.HaHangupCauseCause = "不兼容的目的地[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因指示发送此原因的设备已收到建立呼叫的请求，该呼叫具有无法容纳的低层兼容性，高层兼容性或其他兼容性属性（例如，数据速率）。"

	case "INVALID_MSG_UNSPECIFIED":
		HaHangupV.HaHangupCauseCode = 95
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "INVALID_MSG_UNSPECIFIED"
		HaHangupV.HaHangupCauseCause = "无效消息，未指定[Q.850]"
		HaHangupV.HaHangupCauseDes = "仅当无效消息类中没有其他原因适用时，才使用此原因来报告无效消息事件。"

	case "MANDATORY_IE_MISSING":
		HaHangupV.HaHangupCauseCode = 96
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "MANDATORY_IE_MISSING"
		HaHangupV.HaHangupCauseCause = "必填信息元素丢失[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已接收到一条消息，该消息缺少在处理该消息之前必须存在于该消息中的信息元素。"

	case "MESSAGE_TYPE_NONEXIST":
		HaHangupV.HaHangupCauseCode = 97
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "MESSAGE_TYPE_NONEXIST"
		HaHangupV.HaHangupCauseCause = "消息类型不存在或未实现[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已接收到消息类型不识别的消息，因为这是未定义但未由发送此原因的设备实现的消息。"

	case "WRONG_MESSAGE":
		HaHangupV.HaHangupCauseCode = 98
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "WRONG_MESSAGE"
		HaHangupV.HaHangupCauseCause = "消息与呼叫状态或消息类型不兼容或不存在。[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明发送此原因的设备已收到一条消息，使得该过程不表示这是处于呼叫状态时允许接收的消息，或者已收到指示不兼容呼叫状态的STATUS消息。"

	case "IE_NONEXIST":
		HaHangupV.HaHangupCauseCode = 99
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "IE_NONEXIST"
		HaHangupV.HaHangupCauseCause = "信息元素/参数不存在或未实现[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已收到一条消息，其中包含未识别的信息元素/参数，因为信息元素/参数名称未定义或未定义但未实现发送原因的设备。该原因指示信息元素/参数被丢弃。然而，为了使设备发送原因来处理消息，不需要在消息中存在信息元素。"

	case "INVALID_IE_CONTENTS":
		HaHangupV.HaHangupCauseCode = 100
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "INVALID_IE_CONTENTS"
		HaHangupV.HaHangupCauseCause = "无效的信息元素内容[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已接收并已实现该信息。但是，IE中的一个或多个字段是用发送此原因的设备尚未实现的方式进行编码的。"

	case "WRONG_CALL_STATE":
		HaHangupV.HaHangupCauseCode = 101
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "WRONG_CALL_STATE"
		HaHangupV.HaHangupCauseCause = "消息与呼叫状态不兼容[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明已接收到与呼叫状态不兼容的消息。"

	case "RECOVERY_ON_TIMER_EXPIRE":
		HaHangupV.HaHangupCauseCode = 102
		HaHangupV.HaHangupCauseSipCode = "504"
		HaHangupV.HaHangupCauseName = "RECOVERY_ON_TIMER_EXPIRE"
		HaHangupV.HaHangupCauseCause = "计时器到期后恢复[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明与错误处理过程相关联的计时器已到期，因此该过程已启动。这通常与NAT问题相关。确保在您的ATA中启用了“ NAT映射启用”。如果与NAT不相关，则有时可能与提供商相关，请确保确保另一个出站提供商不能解决问题。当对方发送408呼叫过期时，FreeSWITCH也会返回此消息。"

	case "MANDATORY_IE_LENGTH_ERROR":
		HaHangupV.HaHangupCauseCode = 103
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "MANDATORY_IE_LENGTH_ERROR"
		HaHangupV.HaHangupCauseCause = "参数不存在或未实现-传递（国家使用）[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表示发送此原因的设备已接收到一条消息，该消息包含未识别的参数，因为未定义参数，或者未通过发送此原因的设备来实现参数。原因表明该参数已被忽略。另外，如果发送此原因的设备是中间点，则此原因表明参数已原样传递。"

	case "PROTOCOL_ERROR":
		HaHangupV.HaHangupCauseCode = 111
		HaHangupV.HaHangupCauseSipCode = "501"
		HaHangupV.HaHangupCauseName = "PROTOCOL_ERROR"
		HaHangupV.HaHangupCauseCause = "协议错误，未指定[Q.850]"
		HaHangupV.HaHangupCauseDes = "仅当协议错误类中没有其他原因适用时，才使用此原因来报告协议错误事件。"

	case "ORIGINATOR_CANCEL":
		HaHangupV.HaHangupCauseCode = 487
		HaHangupV.HaHangupCauseSipCode = "487"
		HaHangupV.HaHangupCauseName = "ORIGINATOR_CANCEL"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "CRASH":
		HaHangupV.HaHangupCauseCode = 500
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "CRASH"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "奔溃"

	case "SYSTEM_SHUTDOWN":
		HaHangupV.HaHangupCauseCode = 501
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "SYSTEM_SHUTDOWN"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "系统服务关机"

	case "LOSE_RACE":
		HaHangupV.HaHangupCauseCode = 502
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "LOSE_RACE"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "MANAGER_REQUEST":
		HaHangupV.HaHangupCauseCode = 503
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "MANAGER_REQUEST"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "当您发送api命令使其挂断时，将使用此原因。例如uuid_kill uuid"

	case "BLIND_TRANSFER":
		HaHangupV.HaHangupCauseCode = 600
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "BLIND_TRANSFER"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "ATTENDED_TRANSFER":
		HaHangupV.HaHangupCauseCode = 601
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "ATTENDED_TRANSFER"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "ALLOTTED_TIMEOUT":
		HaHangupV.HaHangupCauseCode = 602
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "ALLOTTED_TIMEOUT"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "此原因意味着服务器取消了呼叫，因为目标通道花费了很长时间才能应答。"

	case "USER_CHALLENGE":
		HaHangupV.HaHangupCauseCode = 603
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "USER_CHALLENGE"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "MEDIA_TIMEOUT":
		HaHangupV.HaHangupCauseCode = 604
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "MEDIA_TIMEOUT"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = ""

	case "PICKED_OFF":
		HaHangupV.HaHangupCauseCode = 605
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "PICKED_OFF"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "此原因意味着呼叫是通过从另一个分机截取来接听的（即从另一个分机拨打** ext_number）。"

	case "USER_NOT_register":
		HaHangupV.HaHangupCauseCode = 606
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "USER_NOT_register"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "这意味着您试图向忘记注册的SIP用户发起呼叫。"

	case "PROGRESS_TIMEOUT":
		HaHangupV.HaHangupCauseCode = 607
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "PROGRESS_TIMEOUT"
		HaHangupV.HaHangupCauseCause = "互通，未指定[Q.850]"
		HaHangupV.HaHangupCauseDes = "此原因表明互通呼叫（通常是对SW56服务的呼叫）已结束。"

	case "GATEWAY_DOWN":
		HaHangupV.HaHangupCauseCode = 609
		HaHangupV.HaHangupCauseSipCode = ""
		HaHangupV.HaHangupCauseName = "GATEWAY_DOWN"
		HaHangupV.HaHangupCauseCause = ""
		HaHangupV.HaHangupCauseDes = "网关已关闭（未回答OPTIONS或SUBSCRIBE）"

	}
	return
}

func ConvertCN(key string) (val string) {
	switch key {
	case "Available":
		val = "空闲状态"
	case "Logged Out":
		val = "注销状态"
	case "On Break":
		val = "小休状态"
	case "Idle":
		val = "队列[空闲]"
	case "Waiting":
		val = "队列[等待]"
	case "In a queue call":
		val = "队列[在队列中呼叫]"
	case "Receiving":
		val = "队列[接听]"
	}
	return
}

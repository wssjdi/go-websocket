package servers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/pkg/setting"
	"strings"
	"time"
)

type Client struct {
	ClientId    string          // 标识ID
	SystemId    string          // 系统ID
	Socket      *websocket.Conn // 用户连接
	ConnectTime uint64          // 首次连接时间
	IsDeleted   bool            // 是否删除或下线
	Notify      bool            // 上下线时是否通知同组内的其他客户端
	UserId      string          // 业务端标识用户ID
	Extend      string          // 扩展字段，用户可以自定义
	GroupList   []string        // 该客户端绑定到的组列表
}

type SendData struct {
	Code int
	Msg  string
	Data *interface{}
}

func NewClient(clientId string, systemId string, notify bool, socket *websocket.Conn) *Client {
	return &Client{
		ClientId:    clientId,
		SystemId:    systemId,
		Socket:      socket,
		ConnectTime: uint64(time.Now().Unix()),
		IsDeleted:   false,
		Notify:      notify,
	}
}

func (c *Client) Read() {
	go func() {
		for {
			messageType, msgBuffer, err := c.Socket.ReadMessage()

			if err != nil {
				log.WithFields(log.Fields{
					"host":        setting.GlobalSetting.LocalHost,
					"port":        setting.CommonSetting.HttpPort,
					"systemId":    c.SystemId,
					"clientId":    c.ClientId,
					"messageType": messageType,
					"message":     string(msgBuffer),
				}).Error("接受到客户端发送的无效消息:" + err.Error())
				if messageType == -1 && websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
					Manager.DisConnect <- c //关闭出错或者断开的连接
					return
				} else if messageType != websocket.PingMessage {
					return
				}
			}

			var msg clientMsg
			if err := json.Unmarshal(msgBuffer, &msg); err == nil {
				//解析成功则处理消息
				handlerClientMsg(c, &msg)
			} else {
				// 解析失败，说明格式不正确，直接忽略
				log.WithFields(log.Fields{
					"host":     setting.GlobalSetting.LocalHost,
					"port":     setting.CommonSetting.HttpPort,
					"systemId": c.SystemId,
					"clientId": c.ClientId,
					"message":  string(msgBuffer),
				}).Error("接受到客户端发送的无效消息:" + err.Error())
			}
		}
	}()
}

func handlerClientMsg(c *Client, msg *clientMsg) {
	// 绑定当前ClientId到组(B2G),需要校验msg的格式,需要判断systemid使用哪一个
	systemId := msg.SystemId
	if len(systemId) == 0 {
		systemId = c.SystemId
	}
	switch strings.ToUpper(msg.Event) {
	case Bind2Group:
		if len(msg.GroupName) > 0 {
			AddClient2Group(systemId, msg.GroupName, c.ClientId, msg.UserId, msg.Extend)
		} else {
			//该操作必传 GroupName,否则忽略
			log.WithFields(log.Fields{
				"event":    msg.Event,
				"host":     setting.GlobalSetting.LocalHost,
				"port":     setting.CommonSetting.HttpPort,
				"systemId": c.SystemId,
				"clientId": c.ClientId,
				"message":  fmt.Sprintf("%+v", msg),
			}).Error("B2G操作,GroupName必传 :")
		}

	case Send2Client, Send2ClientS:
		// 向单个客户端连接发送消息(S2C) ; 同时向多个客户端发送消息(S2M)
		if len(msg.ClientIds) > 0 {
			//该操作必传 ClientIds , 否则忽略
			for _, clientId := range msg.ClientIds {
				//发送信息
				SendMessage2Client(clientId, c.ClientId, retcode.SUCCESS, "success", &msg.Data)
			}
		} else {
			log.WithFields(log.Fields{
				"event":    msg.Event,
				"host":     setting.GlobalSetting.LocalHost,
				"port":     setting.CommonSetting.HttpPort,
				"systemId": c.SystemId,
				"clientId": c.ClientId,
				"message":  fmt.Sprintf("%+v", msg),
			}).Error("S2C、S2M操作,ClientIds必传 :")
		}

	case Send2Group:
		// 同时向群组内所有有效的客户端发送消息(S2G)
		if len(msg.GroupName) > 0 {
			//组发送的同时,如果也传了ClientIds,则使用多发，只发送给指定的客户端
			if len(msg.ClientIds) > 0 {
				for _, clientId := range msg.ClientIds {
					//单个客户端发送信息
					SendMessage2Client(clientId, c.ClientId, retcode.SUCCESS, "success", &msg.Data)
				}
			} else {
				//群发
				SendMessage2Group(systemId, c.ClientId, msg.GroupName, retcode.SUCCESS, "success", &msg.Data)
			}
		} else {
			log.WithFields(log.Fields{
				"event":    msg.Event,
				"host":     setting.GlobalSetting.LocalHost,
				"port":     setting.CommonSetting.HttpPort,
				"systemId": c.SystemId,
				"clientId": c.ClientId,
				"message":  fmt.Sprintf("%+v", msg),
			}).Error("S2G操作,GroupName必传 :")
		}

	case Send2User:
		// TODO: 向拥有相同业务端UserId的客户端发送消息(S2U)
	case Close:
		// 同时向群组内所有有效的客户端发送消息(CLS)
		CloseClient(c.ClientId, systemId)

	default:
		//忽略掉无法识别的消息
		log.WithFields(log.Fields{
			"event":    msg.Event,
			"host":     setting.GlobalSetting.LocalHost,
			"port":     setting.CommonSetting.HttpPort,
			"systemId": c.SystemId,
			"clientId": c.ClientId,
			"message":  fmt.Sprintf("%+v", msg),
		}).Error("接受到客户端发送的无效消息:")
	}
}

type clientMsg struct {
	Event      string   `json:"event" validate:"required"` // 发送消息需要做的操作类型：[绑定到组(B2G)|单发(S2C)|多发(S2M)|群发(S2G)|自发(S2U)|关闭(CLS)]
	SystemId   string   `json:"systemId"`                  // 系统标识，不传则默认使用当前客户端绑定的系统标识，后续可能需要跨系统发送消息
	SendUserId string   `json:"sendUserId"`                // 发送者的clientId，不传则默认使用当前客户端的clientId
	GroupName  string   `json:"groupName"`                 // 群发时候的groupName，无默认值，当event的值为B2G和S2G时必传，否则视为无效消息
	UserId     string   `json:"userId"`                    // 业务端标识用户ID,无默认值，可以在event的值为B2G时绑定一次，后续的操作中可以透传
	Extend     string   `json:"extend"`                    // 业务端扩展字段,无默认值,用户可以自定义,可以在event的值为B2G时绑定一次，后续的操作中可以透传
	ClientIds  []string `json:"clientIds"`                 // 单发或者多发的时候消息接收者的clientId，无默认值，当event的值为S2G时，如clientIds同时不为空，则以clientIds为准，当event的值为S2C或者S2M时必传，否则视为无效消息
	Data       string   `json:"data"`                      // 业务数据，字符串类型，建议使用Json格式，根据各个业务系统需要自定义
}

const (
	// 绑定当前ClientId到组(B2G)
	Bind2Group = "B2G"
	// 向单个客户端连接发送消息(S2C)
	Send2Client = "S2C"
	// 同时向多个客户端发送消息(S2M)
	Send2ClientS = "S2M"
	// 同时向群组内所有有效的客户端发送消息(S2G)
	Send2Group = "S2G"
	// 向拥有相同业务端UserId的客户端发送消息(S2U)
	Send2User = "S2U"
	// 客户端主动向服务器请求关闭连接(CLS)
	Close = "CLS"
)

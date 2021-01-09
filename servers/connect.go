package servers

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/woodylan/go-websocket/api"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/pkg/setting"
	"github.com/woodylan/go-websocket/tools/util"
	"net/http"
	"strings"
)

type Controller struct {
}

type renderData struct {
	ClientId string `json:"clientId"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  setting.CommonSetting.ReadBuffer,
		WriteBufferSize: setting.CommonSetting.WriteBuffer,
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("upgrade error: %v", err)
		http.NotFound(w, r)
		return
	}

	//设置读取消息大小上线
	conn.SetReadLimit(setting.CommonSetting.MaxMessageSize)

	//解析参数
	systemId := r.FormValue("systemId")
	if len(systemId) == 0 {
		_ = Render(conn, "", "", retcode.SystemIdErrCode, "系统ID不能为空", []string{})
		_ = conn.Close()
		return
	}

	clientId := util.GenClientId()

	//上线、下线时是否通知相同GroupName中的其他客户端连接,开启则上线、下线时会通知同组的所有客户端
	notify := false
	if "true" == strings.ToLower(r.FormValue("notify")) {
		notify = true
	}

	clientSocket := NewClient(clientId, systemId, notify, conn)

	Manager.AddClient2SystemClient(systemId, clientSocket)

	//如果有groupName参数,则连接成功之后直接将客户端绑定到对应的组
	groupName := r.FormValue("groupName")
	if len(groupName) > 0 {
		userId := r.FormValue("userId")
		extend := r.FormValue("extend")
		Manager.AddClient2LocalGroup(groupName, clientSocket, userId, extend)
	}

	//读取客户端消息
	clientSocket.Read()

	if err = api.ConnRender(conn, renderData{ClientId: clientId}); err != nil {
		_ = conn.Close()
		return
	}

	// 用户连接事件
	Manager.Connect <- clientSocket
}

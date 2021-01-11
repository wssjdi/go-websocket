package send2clients

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/woodylan/go-websocket/api"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/servers"
	"net/http"
	"strings"
)

type Controller struct {
}

type inputData struct {
	SystemId   string   `json:"systemId"`
	ClientIds  []string `json:"clientIds" validate:"required"`
	SendUserId string   `json:"sendUserId"  validate:"required"`
	Code       int      `json:"code"`
	Msg        string   `json:"msg"`
	Data       string   `json:"data"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	var inputData inputData
	if err := json.NewDecoder(r.Body).Decode(&inputData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := api.Validate(inputData)
	if err != nil {
		api.Render(w, retcode.FAIL, err.Error(), []string{})
		return
	}
	messages := make([]string, len(inputData.ClientIds))
	for _, clientId := range inputData.ClientIds {
		if len(inputData.SendUserId) > 0 && inputData.SendUserId == clientId {
			log.Warnf("过滤掉自己给自己发送的消息~！")
			continue
		}
		//发送信息
		msgId := servers.SendMessage2Client(clientId, inputData.SendUserId, inputData.Code, inputData.Msg, &inputData.Data)
		messages = append(messages, msgId)
	}

	api.Render(w, retcode.SUCCESS, "success", map[string]string{
		"messageId": strings.Join(messages, ","),
	})
	return
}

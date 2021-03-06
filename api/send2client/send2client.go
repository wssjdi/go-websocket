package send2client

import (
	"encoding/json"
	"github.com/woodylan/go-websocket/api"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/servers"
	"net/http"
)

type Controller struct {
}

type inputData struct {
	SystemId   string `json:"systemId"`
	ClientId   string `json:"clientId" validate:"required"`
	SendUserId string `json:"sendUserId"  validate:"required"`
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Data       string `json:"data"`
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

	if inputData.SendUserId == inputData.ClientId {
		api.Render(w, retcode.FAIL, "不允许给自己发送消息", []string{})
		return
	}

	//发送信息
	messageId := servers.SendMessage2Client(inputData.ClientId, inputData.SendUserId, inputData.Code, inputData.Msg, &inputData.Data)

	api.Render(w, retcode.SUCCESS, "success", map[string]string{
		"messageId": messageId,
	})
	return
}

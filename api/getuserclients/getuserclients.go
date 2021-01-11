package getuserclients

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
	SystemId  string      `json:"systemId"`
	GroupName string      `json:"groupName"`
	UserId    string      `json:"userId" validate:"required"`
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
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

	systemId := r.Header.Get("SystemId")
	if len(inputData.SystemId) > 0 {
		systemId = inputData.SystemId
	}

	ret := servers.GetUserList(&systemId, &inputData.GroupName, &inputData.UserId)

	api.Render(w, retcode.SUCCESS, "success", ret)
	return
}

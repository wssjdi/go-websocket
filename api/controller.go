package api

import (
	"encoding/json"
	local "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/woodylan/go-websocket/define/retcode"
	"gopkg.in/go-playground/validator.v9"
	translate "gopkg.in/go-playground/validator.v9/translations/zh"
	"io"
	"net/http"
)

type RetData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ConnRender(conn *websocket.Conn, data interface{}) (err error) {
	err = conn.WriteJSON(RetData{
		Code: retcode.SUCCESS,
		Msg:  "success",
		Data: data,
	})

	return
}

func ConnRenderMsg(conn *websocket.Conn, code int, msg string, data interface{}) (err error) {
	err = conn.WriteJSON(RetData{
		Code: code,
		Msg:  msg,
		Data: data,
	})

	return
}

func Render(w http.ResponseWriter, code int, msg string, data interface{}) (str string) {
	var retData RetData

	retData.Code = code
	retData.Msg = msg
	retData.Data = data

	retJson, _ := json.Marshal(retData)
	str = string(retJson)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.WriteString(w, str)
	return
}

func Validate(inputData interface{}) error {

	validate := validator.New()
	zh := local.New()
	uni := ut.New(zh, zh)
	trans, _ := uni.GetTranslator("zh")

	_ = translate.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(inputData)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(trans))
		}
	}

	return nil
}

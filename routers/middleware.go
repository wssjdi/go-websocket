package routers

import (
	"bytes"
	"encoding/json"
	"github.com/woodylan/go-websocket/api"
	"github.com/woodylan/go-websocket/define"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/pkg/etcd"
	"github.com/woodylan/go-websocket/servers"
	"github.com/woodylan/go-websocket/tools/util"
	"io/ioutil"
	"net/http"
)

type nameSpace struct {
	SystemId string `json:"systemId"`
}

func AccessTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		//检查header是否设置SystemId,header中设置或者在请求体中设置都可以
		systemId := r.Header.Get("SystemId")
		if len(systemId) == 0 {
			bodyBytes, _ := ioutil.ReadAll(r.Body) //请求体读取出来,看有没有传systemId
			//log.Infof("原始请求内容:[%s]" , string(bodyBytes))
			ns := &nameSpace{}
			var err error
			if err = json.Unmarshal(bodyBytes, ns); err == nil {
				systemId = ns.SystemId
			}
			//log.Infof("请求内容反序列化时报错:[%+v]", err)
			//log.Infof("请求内容反序列化为:[%+v]", ns)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) //原始请求内容在放到请求体中，后续的处理要用到
		}

		if len(systemId) == 0 {
			api.Render(w, retcode.SystemIdErrCode, "系统ID不能为空", []string{})
			return
		}

		//判断是否被注册
		if util.IsCluster() {
			resp, err := etcd.Get(define.ETcdPrefixAccountInfo + systemId)
			if err != nil {
				api.Render(w, retcode.FAIL, "etcd服务器错误", []string{})
				return
			}

			if resp.Count == 0 {
				api.Render(w, retcode.SystemIdErrCode, "系统ID无效", []string{})
				return
			}
		} else {
			if _, ok := servers.SystemMap.Load(systemId); !ok {
				api.Render(w, retcode.SystemIdErrCode, "系统ID无效", []string{})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

package main

import (
	"fmt"
	"github.com/woodylan/go-websocket/define"
	"github.com/woodylan/go-websocket/pkg/etcd"
	"github.com/woodylan/go-websocket/pkg/setting"
	"github.com/woodylan/go-websocket/routers"
	"github.com/woodylan/go-websocket/servers"
	"github.com/woodylan/go-websocket/tools/log"
	"github.com/woodylan/go-websocket/tools/util"
	"net"
	"net/http"
)

func init() {
	//读取配置文件,初始化项目配置
	setting.Setup()

	//日志配置初始化
	log.Setup()
}

func main() {

	//初始化RPC服务
	initRPCServer()

	//将服务器地址、端口注册到eTcd中
	registerServer()

	//初始化路由
	routers.Init()

	//启动一个定时器用来发送心跳
	servers.PingTimer()

	fmt.Printf("服务器启动成功，端口号：%s\n", setting.CommonSetting.HttpPort)

	if err := http.ListenAndServe(":"+setting.CommonSetting.HttpPort, nil); err != nil {
		panic(err)
	}
}

func initRPCServer() {
	//如果是集群，则启用RPC进行通讯
	if util.IsCluster() {
		//初始化RPC服务
		servers.InitGRpcServer()
		fmt.Printf("启动RPC，端口号：%s\n", setting.CommonSetting.RPCPort)
	}
}

//eTcd注册发现GRpc服务
func registerServer() {
	if util.IsCluster() {
		//注册租约
		ser, err := etcd.NewServiceReg(setting.EtcdSetting.Endpoints, 5)
		if err != nil {
			panic(err)
		}

		hostPort := net.JoinHostPort(setting.GlobalSetting.LocalHost, setting.CommonSetting.RPCPort)
		//添加key
		err = ser.PutService(define.ETcdServerList+hostPort, hostPort)
		if err != nil {
			panic(err)
		}

		cli, err := etcd.NewClientDis(setting.EtcdSetting.Endpoints)
		if err != nil {
			panic(err)
		}
		_, err = cli.GetService(define.ETcdServerList)
		if err != nil {
			panic(err)
		}
	}
}

package setting

import (
	"flag"
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type commonConf struct {
	HttpPort       string
	RPCPort        string
	Cluster        bool
	CryptoKey      string
	MaxMessageSize int64
	ReadBuffer     int
	WriteBuffer    int
}

var CommonSetting = &commonConf{}

type etcdConf struct {
	Endpoints []string
}

var EtcdSetting = &etcdConf{}

type global struct {
	LocalHost      string //本机内网IP
	ServerList     map[string]string
	ServerListLock sync.RWMutex
}

var GlobalSetting = &global{}

type logConf struct {
	BasePath string //日志文件保存路径
	MaxAge   int    //日志文件保存的时间，单位：天
}

var LogSetting = &logConf{}

var cfg *ini.File

var (
	env     = flag.String("e", "", "The api server run env")
	profile = flag.String("p", "", "The api server run with config file .")
)

func Setup() {
	flag.Parse()
	//启动命令中的profile和env不能同时为空
	if len(*profile) == 0 && len(*env) == 0 {
		log.Fatalf("env type and configure file name is both nil")
	}
	//如果启动命令中没有指定配置文件路径，则使用默认环境下的配置文件
	if len(*profile) == 0 {
		*profile = fmt.Sprintf("conf/app.%s.ini", *env)
	}

	var err error
	cfg, err = ini.Load(*profile)
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse '%s': %v", *profile, err)
	}

	mapTo("common", CommonSetting)
	mapTo("etcd", EtcdSetting)
	mapTo("logfile", LogSetting)

	GlobalSetting = &global{
		LocalHost:  GetIntranetIp(),
		ServerList: make(map[string]string),
	}
}

func Default() {
	CommonSetting = &commonConf{
		HttpPort:       "6000",
		RPCPort:        "7000",
		Cluster:        false,
		CryptoKey:      "axRArfEfJw7V0te6",
		MaxMessageSize: 8192,
		ReadBuffer:     1024,
		WriteBuffer:    1024,
	}

	GlobalSetting = &global{
		LocalHost:  GetIntranetIp(),
		ServerList: make(map[string]string),
	}

	LogSetting = &logConf{
		BasePath: CurrentDirectory(),
		MaxAge:   30,
	}
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}

//获取本机内网IP
func GetIntranetIp() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return ""
}

//获取当前程序运行的文件夹
func CurrentDirectory() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(dir, "\\", "/", -1)
}

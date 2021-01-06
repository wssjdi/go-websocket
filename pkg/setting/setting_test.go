package setting

import (
	"fmt"
	"net"
	"testing"
)

func TestSetting(t *testing.T) {

	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		fmt.Println(fmt.Sprintf("本机地址：%s", addr))
		// 检查ip地址判断是否回环地址
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(fmt.Sprintf("本机非回环地址：%s", ipnet.IP.String()))
			}

		}
	}
}

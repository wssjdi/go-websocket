package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
)

func TestGenUUID(t *testing.T) {
	Convey("生成uuid", t, func() {
		uuid := GenUUID()
		log.Printf("生成新的UUID:[%s]", uuid)
		Convey("验证长度", func() {
			So(len(uuid), ShouldBeGreaterThan, 0)
		})
	})
}

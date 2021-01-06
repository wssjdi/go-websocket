package crypto

import (
	"log"
	"testing"
)

func TestEncrypt(t *testing.T) {
	raw := []byte("password123")
	key := []byte("asdf1234qwer7894")
	str, err := Encrypt(raw, key)
	if err == nil {
		t.Log("suc", str)
	} else {
		t.Fatal("fail", err)
	}
}

func TestDncrypt(t *testing.T) {
	raw := "EsrTlM9RjEDtGYE0ZG9aKHmeQqj1BFD6QFgO5RwfgpXQ/xH7fqKmHVxqRdpXeuXh"
	key := []byte("Adba723b7fe06819")
	str, err := Decrypt(raw, key)
	log.Printf("解密数据为:[ %s ]", str)
	if err == nil {
		t.Log("suc", str)
	} else {
		t.Fatal("fail", err)
	}
}

func Benchmark_CBCEncrypt(b *testing.B) {
	b.StopTimer()
	raw := []byte("146576885")
	key := []byte("gdgghf")

	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(raw, key)
	}
}

func Benchmark_CBCDecrypt(b *testing.B) {
	b.StopTimer()
	raw := "CDl4Uas8ZyaGXaoOhPZ9NLDvcsIkyvvd++TONd8UPZc="
	key := []byte("gdgghf")

	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(raw, key)
	}
}

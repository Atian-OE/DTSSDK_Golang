package dtssdk

import (
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"log"
	"testing"
)

func TestDTSSDKClient(t *testing.T) {
	for index, address := range []string{
		"192.168.0.215",
		"192.168.0.215",
		"192.168.0.215",
	} {
		go func(index int, address string) {
			client := NewDTSClient(address)
			client.CallConnected(func(s string) {
				log.Println("连接成功:", s, index)
				_ = client.CallTempSignalNotify(func(notify *model.TempSignalNotify, err error) {
					log.Println("温度更新", index)
				})
				_ = client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, err error) {
					log.Println("信号更新", index)
				})
			})
			client.CallDisconnected(func(s string) {
				log.Println("断开连接:", s, index)
			})
		}(index, address)

	}
	select {}
}

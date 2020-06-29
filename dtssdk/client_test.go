package dtssdk

import (
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/codec"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"log"
	"net"
	"testing"
	"time"
)

func connect(k, v string) {
	time.Sleep(time.Second)
	client := NewClient(DefaultOptions(k, v))
	client, err := client.Connect()
	if err != nil {
		log.Println("错误 ", err)
	}
	client.CallConnected(func(addr string) {
		time.AfterFunc(time.Minute, func() {
			if k != "设备5" {
				client.Close()
			}
		})
		log.Println(fmt.Sprintf("连接成功:%s!", addr))

		_ = client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, e error) {
			log.Println(client.Options.Id, "温度更新")
		})

		_ = client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
			log.Println(client.Options.Id, "信号更新")
		})

		_ = client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
			log.Println(client.Options.Id, "报警更新")
		})
	})
	client.CallDisconnected(func(addr string) {
		log.Println(client.Options.Id, "断开连接", addr)
		time.AfterFunc(time.Second*10, func() {
			connect(k, v)
		})
	})
}
func TestClient2(t *testing.T) {
	for k, v := range map[string]string{
		"设备1": "192.168.0.215",
		"设备2": "192.168.0.83",
		"设备3": "192.168.0.215",
		"设备4": "192.168.0.83",
		"设备5": "192.168.0.215",
	} {
		connect(k, v)
	}
	time.Sleep(time.Hour)
}

func TestClient(t *testing.T) {
	if c, err := NewClient(Options{
		Id:          "dts",
		Ip:          "192.168.0.215",
		Port:        17083,
		Timeout:     time.Second * 3,
		PingTime:    time.Second * 30,
		ReadBuffer:  5000,
		WriteBuffer: 5000,
	}).CallConnected(func(s string) {
		log.Println("CallConnected")
	}).CallDisconnected(func(s string) {
		log.Println("CallDisconnected")
	}).Connect(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("连接成功", c.addr)
		time.Sleep(time.Second * 5)
		c.Close()
		log.Println("Close")
	}
	time.Sleep(time.Second * 2)
}

func TestName(t *testing.T) {
	dialer := net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	conn, err := dialer.Dial("tcp", "192.168.0.215:17083")
	if err != nil {
		log.Fatal(err)
	}
	if conn, ok := conn.(*net.TCPConn); ok {
		if err := conn.SetKeepAlive(true); err != nil {
			log.Fatal("SetKeepAlive", err)
		}
		if err := conn.SetKeepAlivePeriod(time.Second * 30); err != nil {
			log.Fatal("SetKeepAlivePeriod", err)
		}

		go func() {
			data := make([]byte, 1024)
			if n, err := conn.Read(data); err != nil {
				log.Println(n)
			} else {
				log.Println(n)
			}
		}()

		go func() {
			//req := &model.GetDefenceZoneRequest{}
			//req.Search = ""
			//req.Channel = int32(1)
			b, _ := codec.Encode(&model.HeartBeat{})
			if n, err := conn.Write(b); err != nil {
				log.Println(n)
			}
			time.Sleep(time.Second)
			conn.Close()
		}()
	}
	time.Sleep(time.Hour)
}

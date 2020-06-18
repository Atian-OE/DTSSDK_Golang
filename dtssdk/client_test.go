package dtssdk

import (
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/codec"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"log"
	"net"
	"testing"
	"time"
)

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

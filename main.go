package main

import (
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	client := dtssdk.NewDTSClient("192.168.0.215")
	client.CallConnected(func(addr string) {
		log.Println(fmt.Sprintf("连接成功:%s!", addr))
		if rep, err := client.GetDeviceID(); err == nil {
			if rep.Success {
				log.Println("获取设备ID", rep.DeviceID)
			}
		}

		if rep2, err := client.GetDefenceZone(1, ""); err == nil {
			log.Println("获取防区", len(rep2.Rows))
		}

		if rep3, err := client.CancelSound(); err == nil {
			log.Println("消警", rep3)
		}

		if rep4, err := client.ResetAlarm(); err == nil {
			log.Println("重置警报", rep4)
		}

		_ = client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, e error) {
			log.Println("防区温度更新")
		})

		_ = client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
			log.Println("防区温度信号更新")
		})

		_ = client.CallDeviceEventNotify(func(notify *model.DeviceEventNotify, e error) {
			log.Println("防区事件更新")
		})

		_ = client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
			log.Println("防区报警更新")
		})
	})
	client.CallDisconnected(func(addr string) {
		log.Println(fmt.Sprintf("断开连接:%s!", addr))
	})
	time.AfterFunc(time.Hour*3, func() {
		client.Close()
	})
	<-ch
	client.Close()

	log.Println("quit")
}

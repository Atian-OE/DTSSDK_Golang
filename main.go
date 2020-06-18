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
	log.Println("start")
	client := dtssdk.NewDTSClient("192.168.0.216")
	client.CallConnected(func(addr string) {
		log.Println(fmt.Sprintf("连接成功:%s!", addr))
		rep1, err := client.GetDeviceID()
		log.Println(err)
		log.Println(rep1)

		rep2, err := client.GetDefenceZone(1, "")
		log.Println(err)
		log.Println("GetDefenceZone", rep2, err)

		rep3, err := client.CancelSound()
		log.Println(err)
		log.Println("CancelSound", rep3, err)

		rep4, err := client.ResetAlarm()
		log.Println(err)
		log.Println("ResetAlarm", rep4, err)

		err = client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, e error) {
			log.Println("CallZoneTempNotify" + notify.DeviceID)
		})
		log.Println("CallZoneTempNotify", err)

		err = client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
			log.Println("CallTempSignalNotify" + notify.DeviceID)
		})
		log.Println("CallTempSignalNotify", err)

		err = client.CallDeviceEventNotify(func(notify *model.DeviceEventNotify, e error) {
			log.Println("CallDeviceEventNotify" + notify.DeviceID)
		})
		log.Println("CallDeviceEventNotify", err)

		err = client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
			log.Println("CallZoneAlarmNotify" + notify.DeviceID)
		})
		log.Println("CallZoneAlarmNotify", err)
	})
	client.CallDisconnected(func(addr string) {
		log.Println(fmt.Sprintf("断开连接:%s!", addr))
	})
	time.Sleep(time.Second * 3)

	time.AfterFunc(time.Second*20, func() {
		client.Close()
	})
	<-ch
	client.Close()

	log.Println("quit")
}

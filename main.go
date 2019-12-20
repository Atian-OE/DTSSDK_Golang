package main

import (
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"log"
	"time"
)

func main() {
	fmt.Println(time.Now())
	client := dtssdk.NewDTSClient("192.168.10.215")
	client.SetReconnectTimes(3).
		SetReconnectTime(6).
		SetId("10")
	client.CallConnected(func(addr string) {
		log.Println(fmt.Sprintf("连接成功:%s!", addr))
		log.Println(client.IsReconnecting())
	})
	client.CallDisconnected(func(addr string) {
		log.Println(fmt.Sprintf("断开连接:%s!", addr))
		log.Println(client.IsReconnecting())
	})
	client.CallOntimeout(func(addr string) {
		log.Println(fmt.Sprintf("连接超时:%s!", addr))
		log.Println(client.IsReconnecting())
	})
	time.Sleep(time.Second * 3)

	rep1, err := client.GetDeviceID()
	fmt.Println(err)
	fmt.Println(rep1)

	rep2, err := client.GetDefenceZone(1, "")
	fmt.Println(err)
	fmt.Println("GetDefenceZone", rep2, err)

	rep3, err := client.CancelSound()
	fmt.Println(err)
	fmt.Println("CancelSound", rep3, err)

	rep4, err := client.ResetAlarm()
	fmt.Println(err)
	fmt.Println("ResetAlarm", rep4, err)

	err = client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, e error) {
		fmt.Println("CallZoneTempNotify" + notify.DeviceID)
	})
	fmt.Println("CallZoneTempNotify", err)

	err = client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
		fmt.Println("CallTempSignalNotify" + notify.DeviceID)
	})
	fmt.Println("CallTempSignalNotify", err)

	err = client.CallDeviceEventNotify(func(notify *model.DeviceEventNotify, e error) {
		fmt.Println("CallDeviceEventNotify" + notify.DeviceID)
	})
	fmt.Println("CallDeviceEventNotify", err)

	err = client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
		fmt.Println("CallZoneAlarmNotify" + notify.DeviceID)
	})
	fmt.Println("CallZoneAlarmNotify", err)
	client.CallOnClosed(func() {
		log.Println("over")
	})
	time.Sleep(time.Hour)
	client.Close()
	fmt.Println("quit")
}

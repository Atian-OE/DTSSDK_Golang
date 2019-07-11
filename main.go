package main

import (
	"DTSSDK/dtssdk"
	"DTSSDK/dtssdk/model"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main()  {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	fmt.Println("start")

	client:= dtssdk.NewDTSClient("192.168.0.50")

	client.CallConnected(func() {
		fmt.Println("连接成功!")
	})
	client.CallDisconnected(func() {
		fmt.Println("断开连接!")
	})
	time.Sleep(time.Second*2)

	rep1,err:= client.GetDefenceZone(1,"")
	fmt.Println(err)
	fmt.Println(rep1)


	rep,err:= client.CallZoneTempNotify(func(notify *model.ZoneTempNotify, e error) {
		fmt.Println("CallZoneTempNotify"+notify.DeviceID)
	})
	fmt.Println(err)
	fmt.Println(rep)


	rep,err= client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
		fmt.Println("CallTempSignalNotify"+notify.DeviceID)
	})
	fmt.Println(err)
	fmt.Println(rep)

	rep,err= client.CallDeviceEventNotify(func(notify *model.DeviceEventNotify, e error) {
		fmt.Println("CallDeviceEventNotify"+notify.DeviceID)
	})
	fmt.Println(err)
	fmt.Println(rep)

	rep,err= client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
		fmt.Println("CallZoneAlarmNotify"+notify.DeviceID)
	})
	fmt.Println(err)
	fmt.Println(rep)



	<-ch
	fmt.Println("quit")
}
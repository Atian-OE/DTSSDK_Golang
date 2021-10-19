package main

import (
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {

	fmt.Println(time.Now())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	fmt.Println("start")


	client := dtssdk.NewDTSClient("192.168.0.37")

	client.CallConnected(func(addr string) {
		fmt.Println(fmt.Sprintf("连接成功:%s!", addr))
	})
	client.CallDisconnected(func(addr string) {
		fmt.Println(fmt.Sprintf("断开连接:%s!", addr))
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


	err= client.CallTempSignalNotify(func(notify *model.TempSignalNotify, e error) {
		fmt.Println("CallTempSignalNotify"+notify.DeviceID+" "+ strconv.Itoa( len(notify.Signal)))

	})
	fmt.Println("CallTempSignalNotify", err)

	err = client.CallDeviceEventNotify(func(notify *model.DeviceEventNotify, e error) {
		fmt.Println("CallDeviceEventNotify" + notify.DeviceID)
	})
	fmt.Println("CallDeviceEventNotify", err)

	err = client.CallZoneAlarmNotify(func(notify *model.ZoneAlarmNotify, e error) {
		fmt.Println("CallZoneAlarmNotify" + notify.DeviceID)
	})
	fmt.Println("CallZoneAlarmNotify",err)

	err= client.CallButtonNotify(func(notify *model.ButtonNotify, e error) {
		fmt.Println("CallButtonNotify"+notify.DeviceID)
	})
	fmt.Println("CallButtonNotify",err)



	<-ch
	client.Close()

	fmt.Println("quit")
}

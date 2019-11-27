package dtssdk

import (
	"errors"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/golang/protobuf/proto"
	"net"
)

func (d *DTSSDKClient) tcpHandle(msgId model.MsgID, data []byte, conn net.Conn) {

	var isHandled bool
	d.wait_pack_list.Range(func(key, value interface{}) bool {
		v := value.(*WaitPackStr)
		if v.Key == msgId {
			go (*v.Call)(msgId, data[5:], conn, nil)
			d.wait_pack_list.Delete(key)
			isHandled = true
			return false
		}
		return true
	})
	if isHandled {
		return
	}
	switch msgId {
	case model.MsgID_ConnectID:
		d.connected = true
		go d.SetDeviceRequest()
		if d.connectedAction != nil {
			go d.connectedAction(d.addr)
		}

	case model.MsgID_DisconnectID:
		d.connected = false

		d.wait_pack_list.Range(func(key, value interface{}) bool {
			v := value.(*WaitPackStr)
			go (*v.Call)(0, nil, nil, errors.New("client disconnect"))
			d.wait_pack_list.Delete(key)
			return true
		})

		if d.disconnectedAction != nil {
			go d.disconnectedAction(d.addr)
		}

	case model.MsgID_ZoneTempNotifyID:
		if !d._ZoneTempNotifyEnable {
			return
		}
		reply := model.ZoneTempNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		d._ZoneTempNotify(&reply, err)

	case model.MsgID_ZoneAlarmNotifyID:
		if !d._ZoneAlarmNotifyEnable {
			return
		}
		reply := model.ZoneAlarmNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		d._ZoneAlarmNotify(&reply, err)

	case model.MsgID_DeviceEventNotifyID:
		if !d._FiberStatusNotifyEnable {
			return
		}
		reply := model.DeviceEventNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		d._FiberStatusNotify(&reply, err)

	case model.MsgID_TempSignalNotifyID:
		if !d._TempSignalNotifyEnable {
			return
		}
		reply := model.TempSignalNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		d._TempSignalNotify(&reply, err)
	}

}

//设置设备请求
func (d *DTSSDKClient) SetDeviceRequest() (*model.SetDeviceReply, error) {
	req := &model.SetDeviceRequest{}
	req.ZoneTempNotifyEnable = d._ZoneTempNotifyEnable
	req.ZoneAlarmNotifyEnable = d._ZoneAlarmNotifyEnable
	req.FiberStatusNotifyEnable = d._FiberStatusNotifyEnable
	req.TempSignalNotifyEnable = d._TempSignalNotifyEnable
	err := d.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.SetDeviceReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msgId model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.SetDeviceReply{}
		err = proto.Unmarshal(data, &reply)

		wait <- ReplyStruct{&reply, err}
	}
	d.waitPack(model.MsgID_SetDeviceReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//回调连接到服务器
func (d *DTSSDKClient) CallConnected(call func(string)) {
	d.connectedAction = call
}

//回调断开连接服务器
func (d *DTSSDKClient) CallDisconnected(call func(string)) {
	d.disconnectedAction = call
}

//回调分区温度更新的通知
func (d *DTSSDKClient) CallZoneTempNotify(call func(*model.ZoneTempNotify, error)) error {
	d._ZoneTempNotifyEnable = true
	d._ZoneTempNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !d.connected {
		return errors.New("client not connected")
	}

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区温度更新的通知
func (d *DTSSDKClient) DisableZoneTempNotify() error {
	d._ZoneTempNotifyEnable = false
	d._ZoneTempNotify = nil

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调分区警报更新的通知
func (d *DTSSDKClient) CallZoneAlarmNotify(call func(*model.ZoneAlarmNotify, error)) error {
	d._ZoneAlarmNotifyEnable = true
	d._ZoneAlarmNotify = call
	if call == nil {
		return errors.New("callback func is nil")
	}
	if !d.connected {
		return errors.New("client not connected")
	}

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区警报更新的通知
func (d *DTSSDKClient) DisableZoneAlarmNotify() error {
	d._ZoneAlarmNotifyEnable = false
	d._ZoneAlarmNotify = nil

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调光纤状态更新的通知
func (d *DTSSDKClient) CallDeviceEventNotify(call func(*model.DeviceEventNotify, error)) error {
	d._FiberStatusNotifyEnable = true
	d._FiberStatusNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !d.connected {
		return errors.New("client not connected")
	}

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调光纤状态更新的通知
func (d *DTSSDKClient) DisableDeviceEventNotify() error {
	d._FiberStatusNotifyEnable = false
	d._FiberStatusNotify = nil
	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调温度信号更新的通知
func (d *DTSSDKClient) CallTempSignalNotify(call func(*model.TempSignalNotify, error)) error {
	d._TempSignalNotifyEnable = true
	d._TempSignalNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !d.connected {
		return errors.New("client not connected")
	}

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调温度信号更新的通知
func (d *DTSSDKClient) DisableTempSignalNotify() error {
	d._TempSignalNotifyEnable = false
	d._TempSignalNotify = nil

	reply, err := d.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//获得防区
func (d *DTSSDKClient) GetDefenceZone(chId int, search string) (*model.GetDefenceZoneReply, error) {
	req := &model.GetDefenceZoneRequest{}
	req.Search = search
	req.Channel = int32(chId)

	err := d.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.GetDefenceZoneReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msgId model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.GetDefenceZoneReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	d.waitPack(model.MsgID_GetDefenceZoneReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//获得防区
func (d *DTSSDKClient) GetDeviceID() (*model.GetDeviceIDReply, error) {
	req := &model.GetDeviceIDRequest{}

	err := d.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.GetDeviceIDReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msgId model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.GetDeviceIDReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	d.waitPack(model.MsgID_GetDeviceIDReplyID, &call)

	reply := <-wait
	return reply.rep, reply.err
}

//消音
func (d *DTSSDKClient) CancelSound() (*model.CancelSoundReply, error) {
	req := &model.CancelSoundRequest{}

	err := d.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.CancelSoundReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msgId model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.CancelSoundReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	d.waitPack(model.MsgID_CancelSoundReplyID, &call)

	reply := <-wait
	return reply.rep, reply.err
}

//消音
func (d *DTSSDKClient) ResetAlarm() (*model.ResetAlarmReply, error) {
	req := &model.ResetAlarmRequest{}

	err := d.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.ResetAlarmReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msgId model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.ResetAlarmReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	d.waitPack(model.MsgID_ResetAlarmReplyID, &call)

	reply := <-wait
	return reply.rep, reply.err
}

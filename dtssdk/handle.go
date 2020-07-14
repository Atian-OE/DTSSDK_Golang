package dtssdk

import (
	"errors"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
)

func (self *DTSSDKClient) tcp_handle(msg_id model.MsgID, data []byte, conn net.Conn) {

	var is_handled bool
	self.wait_pack_list.Range(func(key, value interface{}) bool {
		v := value.(*WaitPackStr)
		if v.Key == msg_id {
			go (*v.Call)(msg_id, data[5:], conn, nil)
			self.wait_pack_list.Delete(key)
			is_handled = true
			return false
		}
		return true
	})
	if is_handled {
		return
	}

	switch msg_id {
	case model.MsgID_ConnectID:
		self.connected = true
		go self.SetDeviceRequest()
		log.Println("client connected:", self.addr)
		if self._connected_action != nil {
			go self._connected_action(self.addr)
		}

	case model.MsgID_DisconnectID:
		self.connected = false

		self.wait_pack_list.Range(func(key, value interface{}) bool {
			v := value.(*WaitPackStr)
			go (*v.Call)(0, nil, nil, errors.New("client disconnect"))
			self.wait_pack_list.Delete(key)
			return true
		})

		if self._disconnected_action != nil {
			go self._disconnected_action(self.addr)
		}

	case model.MsgID_ZoneTempNotifyID:
		if !self._ZoneTempNotifyEnable {
			return
		}
		reply := model.ZoneTempNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		self._ZoneTempNotify(&reply, err)

	case model.MsgID_ZoneAlarmNotifyID:
		if !self._ZoneAlarmNotifyEnable {
			return
		}
		reply := model.ZoneAlarmNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		self._ZoneAlarmNotify(&reply, err)

	case model.MsgID_DeviceEventNotifyID:
		if !self._FiberStatusNotifyEnable {
			return
		}
		reply := model.DeviceEventNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		self._FiberStatusNotify(&reply, err)

	case model.MsgID_TempSignalNotifyID:
		if !self._TempSignalNotifyEnable {
			return
		}
		reply := model.TempSignalNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		self._TempSignalNotify(&reply, err)
	}

}

//设置设备请求
func (self *DTSSDKClient) SetDeviceRequest() (*model.SetDeviceReply, error) {
	req := &model.SetDeviceRequest{}
	req.ZoneTempNotifyEnable = self._ZoneTempNotifyEnable
	req.ZoneAlarmNotifyEnable = self._ZoneAlarmNotifyEnable
	req.FiberStatusNotifyEnable = self._FiberStatusNotifyEnable
	req.TempSignalNotifyEnable = self._TempSignalNotifyEnable
	err := self.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.SetDeviceReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msg_id model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.SetDeviceReply{}
		err = proto.Unmarshal(data, &reply)

		wait <- ReplyStruct{&reply, err}
	}
	self.wait_pack(model.MsgID_SetDeviceReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//回调连接到服务器
func (self *DTSSDKClient) CallConnected(call func(string)) {
	self._connected_action = call
}

//回调断开连接服务器
func (self *DTSSDKClient) CallDisconnected(call func(string)) {
	self._disconnected_action = call
}

//回调分区温度更新的通知
func (self *DTSSDKClient) CallZoneTempNotify(call func(*model.ZoneTempNotify, error)) error {
	self._ZoneTempNotifyEnable = true
	self._ZoneTempNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !self.connected {
		return errors.New("client not connected")
	}

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区温度更新的通知
func (self *DTSSDKClient) DisableZoneTempNotify() error {
	self._ZoneTempNotifyEnable = false
	self._ZoneTempNotify = nil

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调分区警报更新的通知
func (self *DTSSDKClient) CallZoneAlarmNotify(call func(*model.ZoneAlarmNotify, error)) error {
	self._ZoneAlarmNotifyEnable = true
	self._ZoneAlarmNotify = call
	if call == nil {
		return errors.New("callback func is nil")
	}
	if !self.connected {
		return errors.New("client not connected")
	}

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区警报更新的通知
func (self *DTSSDKClient) DisableZoneAlarmNotify() error {
	self._ZoneAlarmNotifyEnable = false
	self._ZoneAlarmNotify = nil

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调光纤状态更新的通知
func (self *DTSSDKClient) CallDeviceEventNotify(call func(*model.DeviceEventNotify, error)) error {
	self._FiberStatusNotifyEnable = true
	self._FiberStatusNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !self.connected {
		return errors.New("client not connected")
	}

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调光纤状态更新的通知
func (self *DTSSDKClient) DisableDeviceEventNotify() error {
	self._FiberStatusNotifyEnable = false
	self._FiberStatusNotify = nil
	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调温度信号更新的通知
func (self *DTSSDKClient) CallTempSignalNotify(call func(*model.TempSignalNotify, error)) error {
	self._TempSignalNotifyEnable = true
	self._TempSignalNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !self.connected {
		return errors.New("client not connected")
	}

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调温度信号更新的通知
func (self *DTSSDKClient) DisableTempSignalNotify() error {
	self._TempSignalNotifyEnable = false
	self._TempSignalNotify = nil

	reply, err := self.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//获得防区
func (self *DTSSDKClient) GetDefenceZone(ch_id int, search string) (*model.GetDefenceZoneReply, error) {
	req := &model.GetDefenceZoneRequest{}
	req.Search = search
	req.Channel = int32(ch_id)

	err := self.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.GetDefenceZoneReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msg_id model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.GetDefenceZoneReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	self.wait_pack(model.MsgID_GetDefenceZoneReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//获得防区
func (self *DTSSDKClient) GetDeviceID() (*model.GetDeviceIDReply, error) {
	req := &model.GetDeviceIDRequest{}

	err := self.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.GetDeviceIDReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msg_id model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.GetDeviceIDReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	self.wait_pack(model.MsgID_GetDeviceIDReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

//消音
func (self *DTSSDKClient) CancelSound() (*model.CancelSoundReply, error) {
	req := &model.CancelSoundRequest{}

	err := self.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.CancelSoundReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msg_id model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.CancelSoundReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	self.wait_pack(model.MsgID_CancelSoundReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

//消音
func (self *DTSSDKClient) ResetAlarm() (*model.ResetAlarmReply, error) {
	req := &model.ResetAlarmRequest{}

	err := self.Send(req)
	if err != nil {
		return nil, err
	}

	type ReplyStruct struct {
		rep *model.ResetAlarmReply
		err error
	}

	wait := make(chan ReplyStruct)

	call := func(msg_id model.MsgID, data []byte, conn net.Conn, err error) {
		if err != nil {
			wait <- ReplyStruct{nil, err}
			return
		}
		reply := model.ResetAlarmReply{}
		err = proto.Unmarshal(data, &reply)
		wait <- ReplyStruct{&reply, err}
	}

	self.wait_pack(model.MsgID_ResetAlarmReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

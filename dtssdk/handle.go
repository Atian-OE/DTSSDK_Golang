package dtssdk

import (
	"errors"
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
)

func (c *Client) tcpHandle(msgId model.MsgID, data []byte, conn net.Conn) {

	var isHandled bool
	c.waitPackList.Range(func(key, value interface{}) bool {
		v := value.(*WaitPackStr)
		if v.Key == msgId {
			go (*v.Call)(msgId, data[5:], conn, nil)
			c.waitPackList.Delete(key)
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
		c.connected = true
		go c.SetDeviceRequest()
		log.Println(fmt.Sprintf("dts客户端已经连接到服务端[ %s ]", c.addr))
		if c.connectedAction != nil {
			go c.connectedAction(c.addr)
		}

	case model.MsgID_DisconnectID:
		c.connected = false

		c.waitPackList.Range(func(key, value interface{}) bool {
			v := value.(*WaitPackStr)
			go (*v.Call)(0, nil, nil, errors.New("client disconnect"))
			c.waitPackList.Delete(key)
			return true
		})

		if c.disconnectedAction != nil {
			go c.disconnectedAction(c.addr)
		}

	case model.MsgID_ZoneTempNotifyID:
		if !c._ZoneTempNotifyEnable {
			return
		}
		reply := model.ZoneTempNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		c._ZoneTempNotify(&reply, err)

	case model.MsgID_ZoneAlarmNotifyID:
		if !c._ZoneAlarmNotifyEnable {
			return
		}
		reply := model.ZoneAlarmNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		c._ZoneAlarmNotify(&reply, err)

	case model.MsgID_DeviceEventNotifyID:
		if !c._FiberStatusNotifyEnable {
			return
		}
		reply := model.DeviceEventNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		c._FiberStatusNotify(&reply, err)

	case model.MsgID_TempSignalNotifyID:
		if !c._TempSignalNotifyEnable {
			return
		}

		reply := model.TempSignalNotify{}
		err := proto.Unmarshal(data[5:], &reply)
		c._TempSignalNotify(&reply, err)

	case model.MsgID_ButtonNotifyID:
		if(!c._ButtonNotifyEnable){
			return
		}
		reply:=model.ButtonNotify{}
		err:=proto.Unmarshal(data[5:], &reply)
		c._ButtonNotify(&reply,err)

	}

}

//设置设备请求

func (c *Client) SetDeviceRequest() (*model.SetDeviceReply, error) {
	req := &model.SetDeviceRequest{}
	req.ZoneTempNotifyEnable = c._ZoneTempNotifyEnable
	req.ZoneAlarmNotifyEnable = c._ZoneAlarmNotifyEnable
	req.FiberStatusNotifyEnable = c._FiberStatusNotifyEnable
	req.TempSignalNotifyEnable = c._TempSignalNotifyEnable
	req.ButtonNotifyEnable=c._ButtonNotifyEnable
	err := c.Send(req)
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
	c.waitPack(model.MsgID_SetDeviceReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//回调连接到服务器
func (c *Client) CallConnected(call func(string)) {
	c.connectedAction = call
}

//回调断开连接服务器
func (c *Client) CallDisconnected(call func(string)) {
	c.disconnectedAction = call
}

//回调分区温度更新的通知
func (c *Client) CallZoneTempNotify(call func(*model.ZoneTempNotify, error)) error {
	c._ZoneTempNotifyEnable = true
	c._ZoneTempNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !c.connected {
		return errors.New("client not connected")
	}

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区温度更新的通知
func (c *Client) DisableZoneTempNotify() error {
	c._ZoneTempNotifyEnable = false
	c._ZoneTempNotify = nil

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调分区警报更新的通知
func (c *Client) CallZoneAlarmNotify(call func(*model.ZoneAlarmNotify, error)) error {
	c._ZoneAlarmNotifyEnable = true
	c._ZoneAlarmNotify = call
	if call == nil {
		return errors.New("callback func is nil")
	}
	if !c.connected {
		return errors.New("client not connected")
	}

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调分区警报更新的通知
func (c *Client) DisableZoneAlarmNotify() error {
	c._ZoneAlarmNotifyEnable = false
	c._ZoneAlarmNotify = nil

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调光纤状态更新的通知
func (c *Client) CallDeviceEventNotify(call func(*model.DeviceEventNotify, error)) error {
	c._FiberStatusNotifyEnable = true
	c._FiberStatusNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !c.connected {
		return errors.New("client not connected")
	}

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调光纤状态更新的通知
func (c *Client) DisableDeviceEventNotify() error {
	c._FiberStatusNotifyEnable = false
	c._FiberStatusNotify = nil
	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调温度信号更新的通知
func (c *Client) CallTempSignalNotify(call func(*model.TempSignalNotify, error)) error {
	c._TempSignalNotifyEnable = true
	c._TempSignalNotify = call

	if call == nil {
		return errors.New("callback func is nil")
	}
	if !c.connected {
		return errors.New("client not connected")
	}

	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调温度信号更新的通知

func (self*Client)DisableButtonNotify() error {
	self._ButtonNotifyEnable =false
	self._ButtonNotify =nil

	reply,err:=self.SetDeviceRequest()
	if(err!=nil){
		return err
	}
	if(!reply.Success){
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//回调按钮触发的通知
func (self*Client)CallButtonNotify(call func(*model.ButtonNotify,error)) error {
	self._ButtonNotifyEnable =true
	self._ButtonNotify =call

	if(call==nil){
		return errors.New("callback func is nil")
	}
	if(!self.connected){
		return errors.New("client not connected")
	}


	reply,err:=self.SetDeviceRequest()
	if(err!=nil){
		return err
	}
	if(!reply.Success){
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//禁用回调温度信号更新的通知
func (c *Client)DisableTempSignalNotify() error {
	c._TempSignalNotifyEnable =false
	c._TempSignalNotify =nil


	reply, err := c.SetDeviceRequest()
	if err != nil {
		return err
	}
	if !reply.Success {
		return errors.New(reply.ErrMsg)
	}
	return nil
}

//获得防区
func (c *Client) GetDefenceZone(ch_id int, search string) (*model.GetDefenceZoneReply, error) {
	req := &model.GetDefenceZoneRequest{}
	req.Search = search
	req.Channel = int32(ch_id)

	err := c.Send(req)
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

	c.waitPack(model.MsgID_GetDefenceZoneReplyID, &call)

	reply := <-wait

	return reply.rep, reply.err
}

//获得防区
func (c *Client) GetDeviceID() (*model.GetDeviceIDReply, error) {
	req := &model.GetDeviceIDRequest{}

	err := c.Send(req)
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

	c.waitPack(model.MsgID_GetDeviceIDReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

//消音
func (c *Client) CancelSound() (*model.CancelSoundReply, error) {
	req := &model.CancelSoundRequest{}

	err := c.Send(req)
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

	c.waitPack(model.MsgID_CancelSoundReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

//消音
func (c *Client) ResetAlarm() (*model.ResetAlarmReply, error) {
	req := &model.ResetAlarmRequest{}

	err := c.Send(req)
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

	c.waitPack(model.MsgID_ResetAlarmReplyID, &call)

	reply := <-wait

	if err != nil {
		return nil, err
	}

	return reply.rep, reply.err
}

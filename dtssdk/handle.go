package dtssdk

import (
	"errors"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/golang/protobuf/proto"
	"net"
)

func (self*DTSSDKClient)tcp_handle(msg_id model.MsgID, data []byte,conn net.Conn)  {

	self.wait_pack_list_mu.Lock()
	defer self.wait_pack_list_mu.Unlock()
	for l:=self.wait_pack_list.Front();l!=nil;l=l.Next(){
		v:=l.Value.(WaitPackStr)
		//fmt.Println(v.Key.String(),msg_type.String())
		if(v.Key==msg_id){
			go (*v.Call)(msg_id,data[5:],conn,nil)
			self.wait_pack_list.Remove(l)
			return
		}
	}

	switch (msg_id) {
	case model.MsgID_ConnectID:
		self.connected=true
		go self.SetDeviceRequest()
		if(self._connected_action!=nil){
			self._connected_action(self.addr)
		}

	case model.MsgID_DisconnectID:
		self.connected=false

		for l:=self.wait_pack_list.Front();l!=nil;l=l.Next(){
			v:=l.Value.(WaitPackStr)
				go (*v.Call)(0,nil,nil,errors.New("client disconnect"))
				self.wait_pack_list.Remove(l)
		}

		if(self._disconnected_action!=nil){
			self._disconnected_action(self.addr)
		}

	case model.MsgID_ZoneTempNotifyID:
		if(!self._ZoneTempNotifyEnable){
			return
		}
		reply:=model.ZoneTempNotify{}
		err:=proto.Unmarshal(data[5:], &reply)
		self._ZoneTempNotify(&reply,err)

	case model.MsgID_ZoneAlarmNotifyID:
		if(!self._ZoneAlarmNotifyEnable){
			return
		}
		reply:=model.ZoneAlarmNotify{}
		err:=proto.Unmarshal(data[5:], &reply)
		self._ZoneAlarmNotify(&reply,err)

	case model.MsgID_DeviceEventNotifyID:
		if(!self._FiberStatusNotifyEnable){
			return
		}
		reply:=model.DeviceEventNotify{}
		err:=proto.Unmarshal(data[5:], &reply)
		self._FiberStatusNotify(&reply,err)

	case model.MsgID_TempSignalNotifyID:
		if(!self._TempSignalNotifyEnable){
			return
		}
		reply:=model.TempSignalNotify{}
		err:=proto.Unmarshal(data[5:], &reply)
		self._TempSignalNotify(&reply,err)
	}

}


//设置设备请求
func (self*DTSSDKClient)SetDeviceRequest() (*model.SetDeviceReply,error) {
	req:=&model.SetDeviceRequest{}
	req.ZoneTempNotifyEnable=self._ZoneTempNotifyEnable
	req.ZoneAlarmNotifyEnable=self._ZoneAlarmNotifyEnable
	req.FiberStatusNotifyEnable=self._FiberStatusNotifyEnable
	req.TempSignalNotifyEnable=self._TempSignalNotifyEnable
	err:=self.Send(req)
	if(err!=nil){
		return nil,err
	}

	//time.After(time.Second*3)
	type ReplyStruct struct {
		rep *model.SetDeviceReply
		err error
	}

	wait:=make(chan ReplyStruct)

	call:= func(msg_id model.MsgID,data []byte,conn net.Conn,err error) {
		if(err!=nil){
			wait<-ReplyStruct{nil,err}
			return
		}
		reply:=model.SetDeviceReply{}
		err=proto.Unmarshal(data, &reply)

		wait<-ReplyStruct{&reply,err}
	}
	self.WaitPack(model.MsgID_SetDeviceReplyID, &call)
	reply:=<-wait


	return reply.rep,reply.err
}

//回调连接到服务器
func (self*DTSSDKClient)CallConnected(call func(string))  {
	self._connected_action=call
}

//回调断开连接服务器
func (self*DTSSDKClient)CallDisconnected(call func(string))  {
	self._disconnected_action=call
}

//回调分区温度更新的通知
func (self*DTSSDKClient)CallZoneTempNotify(call func(*model.ZoneTempNotify,error)) (*model.SetDeviceReply, error) {
	self._ZoneTempNotifyEnable =true
	self._ZoneTempNotify =call

	if(call==nil){
		return nil, errors.New("callback func is nil")
	}
	if(!self.connected){
		return nil, errors.New("client not connected")
	}

	return self.SetDeviceRequest()
}

//禁用回调分区温度更新的通知
func (self*DTSSDKClient)DisableZoneTempNotify() (*model.SetDeviceReply, error) {
	self._ZoneTempNotifyEnable =false
	self._ZoneTempNotify =nil

	return self.SetDeviceRequest()
}


//回调分区警报更新的通知
func (self*DTSSDKClient)CallZoneAlarmNotify(call func(*model.ZoneAlarmNotify,error)) (*model.SetDeviceReply, error) {
	self._ZoneAlarmNotifyEnable =true
	self._ZoneAlarmNotify =call
	if(call==nil){
		return nil, errors.New("callback func is nil")
	}
	if(!self.connected){
		return nil, errors.New("client not connected")
	}

	return self.SetDeviceRequest()
}

//禁用回调分区警报更新的通知
func (self*DTSSDKClient)DisableZoneAlarmNotify() (*model.SetDeviceReply, error) {
	self._ZoneAlarmNotifyEnable =false
	self._ZoneAlarmNotify =nil

	return self.SetDeviceRequest()
}


//回调光纤状态更新的通知
func (self*DTSSDKClient)CallDeviceEventNotify(call func(*model.DeviceEventNotify,error)) (*model.SetDeviceReply, error) {
	self._FiberStatusNotifyEnable =true
	self._FiberStatusNotify =call

	if(call==nil){
		return nil, errors.New("callback func is nil")
	}
	if(!self.connected){
		return nil, errors.New("client not connected")
	}

	return self.SetDeviceRequest()
}

//禁用回调光纤状态更新的通知
func (self*DTSSDKClient)DisableDeviceEventNotify() (*model.SetDeviceReply, error) {
	self._FiberStatusNotifyEnable =false
	self._FiberStatusNotify =nil

	return self.SetDeviceRequest()
}


//回调温度信号更新的通知
func (self*DTSSDKClient)CallTempSignalNotify(call func(*model.TempSignalNotify,error)) (*model.SetDeviceReply, error) {
	self._TempSignalNotifyEnable =true
	self._TempSignalNotify =call

	if(call==nil){
		return nil, errors.New("callback func is nil")
	}
	if(!self.connected){
		return nil, errors.New("client not connected")
	}


	return self.SetDeviceRequest()
}

//禁用回调温度信号更新的通知
func (self*DTSSDKClient)DisableTempSignalNotify() (*model.SetDeviceReply, error) {
	self._TempSignalNotifyEnable =false
	self._TempSignalNotify =nil

	return self.SetDeviceRequest()
}


//获得防区
func (self*DTSSDKClient)GetDefenceZone(ch_id int,search string) (*model.GetDefenceZoneReply,error) {
	req:=&model.GetDefenceZoneRequest{}
	req.Search=search
	req.Channel=int32(ch_id)

	err:=self.Send(req)
	if(err!=nil){
		return nil,err
	}

	type ReplyStruct struct {
		rep *model.GetDefenceZoneReply
		err error
	}

	wait:=make(chan ReplyStruct)
	//wait:=make(chan model.GetDefenceZoneReply)

	call:= func(msg_id model.MsgID,data []byte,conn net.Conn,err error) {
		if(err!=nil){
			wait<-ReplyStruct{nil,err}
			return
		}
		reply:=model.GetDefenceZoneReply{}
		err=proto.Unmarshal(data, &reply)
		wait<-ReplyStruct{&reply,err}
	}

	self.WaitPack(model.MsgID_GetDefenceZoneReplyID, &call)
	reply:=<-wait

	return reply.rep,reply.err
}

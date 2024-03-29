package codec

import (
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/utils"
	"github.com/golang/protobuf/proto"
)

//length 4 msg_id 1  包体 。。。
func Encode(msg_obj interface{}) ([]byte, error) {
	data, err := proto.Marshal(msg_obj.(proto.Message))
	if err != nil {
		return nil, err
	}
	cache := make([]byte, len(data)+5)
	length, _ := utils.IntToBytes(int64(len(data)), 4)
	copy(cache, length)
	switch msg_obj.(type) {
	case *model.GetDefenceZoneRequest:
		cache[4] = byte(model.MsgID_GetDefenceZoneRequestID)
	case *model.GetDefenceZoneReply:
		cache[4] = byte(model.MsgID_GetDefenceZoneReplyID)
	case *model.SetDeviceRequest:
		cache[4] = byte(model.MsgID_SetDeviceRequestID)
	case *model.SetDeviceReply:
		cache[4] = byte(model.MsgID_SetDeviceReplyID)
	case *model.GetDeviceIDRequest:
		cache[4] = byte(model.MsgID_GetDeviceIDRequestID)
	case *model.ZoneTempNotify:
		cache[4] = byte(model.MsgID_ZoneTempNotifyID)
	case *model.ZoneAlarmNotify:
		cache[4] = byte(model.MsgID_ZoneAlarmNotifyID)
	case *model.DeviceEventNotify:
		cache[4] = byte(model.MsgID_DeviceEventNotifyID)
	case *model.TempSignalNotify:

		cache[4]=byte(model.MsgID_TempSignalNotifyID)
	case *model.ButtonNotify:
		cache[4]=byte(model.MsgID_ButtonNotifyID)

	case *model.CancelSoundReply:
		cache[4] = byte(model.MsgID_CancelSoundReplyID)
	case *model.CancelSoundRequest:
		cache[4] = byte(model.MsgID_CancelSoundRequestID)
	case *model.ResetAlarmRequest:
		cache[4] = byte(model.MsgID_ResetAlarmRequestID)
	case *model.ResetAlarmReply:
		cache[4] = byte(model.MsgID_ResetAlarmReplyID)
	case *model.HeartBeat:
		cache[4] = byte(model.MsgID_HeartBeatID)
	}
	copy(cache[5:], data)

	return cache, err
}

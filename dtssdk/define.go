package dtssdk

import (
	"errors"
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"net"
	"sync"
	"time"
)

type WaitPackStr struct {
	Key     model.MsgID
	Timeout int64 //毫秒
	Call    *func(model.MsgID, []byte, net.Conn, error)
}

// Deprecated
type DTSSDKClient = Client

type Client struct {
	sess                  *net.TCPConn
	id                    string
	connected             bool
	count                 int
	reconnecting          bool             //正在重新连接的标志
	waitPackList          *sync.Map        //等待这个包回传
	waitPackTimeoutTicker *time.Ticker     //等待回传的回调 会在 3秒后 自动删除
	waitPackTimeoutOver   chan interface{} //关闭自动删除
	heartBeatTicker       *time.Ticker     //心跳包的发送
	heartBeatTickerOver   chan interface{} //关闭心跳

	reconnectTime       time.Duration    //重连时间
	reconnectTicker     *time.Ticker     //自动连接
	reconnectTickerOver chan interface{} //关闭自动连接

	reconnectTimes           int
	addr                     string                                //地址
	port                     int                                   //端口 默认17083
	connectedAction          func(string)                          //连接到服务器的回调
	closedAction             func()                                //连接到服务器的回调
	disconnectedAction       func(string)                          //断开连接到服务器的回调
	timeoutAction            func(string)                          //连接超时回调
	_ZoneTempNotifyEnable    bool                                  //接收分区温度更新的通知
	_ZoneTempNotify          func(*model.ZoneTempNotify, error)    //分区温度更新
	_ZoneAlarmNotifyEnable   bool                                  //接收温度警报的通知
	_ZoneAlarmNotify         func(*model.ZoneAlarmNotify, error)   //分区警报通知
	_FiberStatusNotifyEnable bool                                  //接收设备状态改变的通知
	_FiberStatusNotify       func(*model.DeviceEventNotify, error) //设备状态通知
	_TempSignalNotifyEnable  bool                                  //接收设备温度信号的通知
	_TempSignalNotify        func(*model.TempSignalNotify, error)  //设备状态通知
}

var (
	ErrClientNotConnect = func(ip string) error {
		return errors.New(fmt.Sprintf("DTS客户端未连接服务端[ %s ]", ip))
	}

	ErrCallback = func(name string) error {
		return errors.New(fmt.Sprintf("DTS客户端[ %s接口 ]未添加回调函数", name))
	}
)

func (c *Client) IsReconnecting() bool {
	return c.reconnecting
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) SetId(id string) *Client {
	c.id = id
	return c
}

func (c *Client) Port() int {
	return c.port
}

func (c *Client) SetPort(port int) *Client {
	c.port = port
	return c
}

func (c *Client) ReconnectTimes() int {
	return c.reconnectTimes
}

func (c *Client) SetReconnectTimes(reconnectTimes int) *Client {
	c.reconnectTimes = reconnectTimes
	return c
}

func (c *Client) ReconnectTime() time.Duration {
	if c.reconnectTime == 0 {
		c.reconnectTime = 10
	}
	return c.reconnectTime
}

func (c *Client) SetReconnectTime(reconnectTime time.Duration) *Client {
	c.reconnectTime = reconnectTime
	return c
}

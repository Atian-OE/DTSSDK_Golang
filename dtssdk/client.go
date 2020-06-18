package dtssdk

import (
	"bytes"
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/codec"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/utils"
	"github.com/kataras/iris/core/errors"
	"log"
	"net"
	"sync"
	"time"
)

type WaitPackStr struct {
	Key     model.MsgID
	Timeout int64 //毫秒
	Call    *func(model.MsgID, []byte, net.Conn, error)
}

// Deprecated: users should Client instead
type DTSSDKClient Client

type Client struct {
	Options
	sess                  *net.TCPConn
	connected             bool
	waitPackList          *sync.Map        //等待这个包回传
	waitPackTimeoutTicker *time.Ticker     //等待回传的回调 会在 3秒后 自动删除
	waitPackTimeoutOver   chan interface{} //关闭自动删除
	heartBeatTicker       *time.Ticker     //心跳包的发送
	heartBeatTickerOver   chan interface{} //关闭心跳

	reconnectTicker     *time.Ticker     //自动连接
	reconnectTickerOver chan interface{} //关闭自动连接

	addr string //地址

	connectedAction          func(string)                          //连接到服务器的回调
	disconnectedAction       func(string)                          //断开连接到服务器的回调
	_ZoneTempNotifyEnable    bool                                  //接收分区温度更新的通知
	_ZoneTempNotify          func(*model.ZoneTempNotify, error)    //分区温度更新
	_ZoneAlarmNotifyEnable   bool                                  //接收温度警报的通知
	_ZoneAlarmNotify         func(*model.ZoneAlarmNotify, error)   //分区警报通知
	_FiberStatusNotifyEnable bool                                  //接收设备状态改变的通知
	_FiberStatusNotify       func(*model.DeviceEventNotify, error) //设备状态通知
	_TempSignalNotifyEnable  bool                                  //接收设备温度信号的通知
	_TempSignalNotify        func(*model.TempSignalNotify, error)  //设备状态通知
}

func NewClient(o Options) *Client {
	return &Client{Options: o}
}

func NewDTSClient(ip string) *Client {
	conn := &Client{
		addr: fmt.Sprintf("%s:%d", ip, 17083),
	}
	conn.init()
	return conn
}

func (c *Client) init() {
	c.reconnectTicker = time.NewTicker(time.Second * 10)
	c.reconnectTickerOver = make(chan interface{})
	go c.reconnect()
}

func (c *Client) Connect() (*Client, error) {
	if c.Ip == "" {
		return c, errors.New("ip need")
	}
	c.addr = fmt.Sprintf("%s:%d", c.Ip, 17083)
	if c.Timeout == 0 {
		c.Timeout = time.Second * 3
	}
	return c, c.connect()
}

func (c *Client) connect() error {
	if c.connected {
		return nil
	}
	conn, err := net.DialTimeout("tcp", c.addr, c.Timeout)
	if err != nil {
		log.Println("连接服务器失败!")
		return err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		c.sess = tcpConn
	} else {
		return errors.New("")
	}

	if err := c.sess.SetWriteBuffer(5000); err != nil {
		return err
	}
	if err := c.sess.SetReadBuffer(5000); err != nil {
		return err
	}
	go c.waitPackTimeout()
	go c.heartBeat()
	go c.clientHandle(c.sess)
	return nil
}

func (c *Client) reconnect() {
	c.connected = false
	c.connect()
	for {
		select {
		case <-c.reconnectTicker.C:
			c.connect()

		case <-c.reconnectTickerOver:
			c.reconnectTicker.Stop()
			return
		}
	}
}

//心跳
func (c *Client) heartBeat() {
	c.heartBeatTicker = time.NewTicker(time.Second * 5)
	c.heartBeatTickerOver = make(chan interface{})
	for {
		select {
		case <-c.heartBeatTicker.C:
			if c.connected {
				log.Println("心跳")
				b, _ := codec.Encode(&model.HeartBeat{})
				if _, err := c.sess.Write(b); err != nil {
					log.Println("发送失败")
					c.Close()
				}
			}

		case <-c.heartBeatTickerOver:
			c.heartBeatTicker.Stop()
			return
		}
	}
}

//超时删除回调
func (c *Client) waitPackTimeout() {
	c.waitPackList = new(sync.Map)
	c.waitPackTimeoutTicker = time.NewTicker(time.Millisecond * 500)
	c.waitPackTimeoutOver = make(chan interface{})
	for {
		select {
		case <-c.waitPackTimeoutTicker.C:
			c.waitPackList.Range(func(key, value interface{}) bool {
				v := value.(*WaitPackStr)
				v.Timeout -= 500
				if v.Timeout <= 0 {
					go (*v.Call)(0, nil, nil, errors.New("callback timeout"))
					c.waitPackList.Delete(key)
				}
				return true
			})

		case <-c.waitPackTimeoutOver:
			c.waitPackTimeoutTicker.Stop()
			return
		}
	}
}

func (c *Client) clientHandle(conn net.Conn) {
	c.tcpHandle(model.MsgID_ConnectID, nil, conn)
	buf := make([]byte, 1024)
	var cache bytes.Buffer
	for {
		n, err := conn.Read(buf)
		//加上上一次的缓存
		if err != nil {
			break
		}

		cache.Write(buf[:n])
		for {
			if c.unpack(&cache, conn) {
				break
			}
		}
	}
}

// true 处理完成 false 循环继续处理
func (c *Client) unpack(cache *bytes.Buffer, conn net.Conn) bool {
	if cache.Len() < 5 {
		return true
	}
	buf := cache.Bytes()
	pkgSize := utils.ByteToInt2(buf[:4])
	//长度不够
	if pkgSize > len(buf)-5 {
		return true
	}

	cmd := buf[4]
	c.tcpHandle(model.MsgID(cmd), buf[:pkgSize+5], conn)
	cache.Reset()
	cache.Write(buf[5+pkgSize:])

	return false
}

//这个包会由这个回调接受
func (c *Client) waitPack(msgId model.MsgID, call *func(model.MsgID, []byte, net.Conn, error)) {
	c.waitPackList.Store(call, &WaitPackStr{Key: msgId, Timeout: 10000, Call: call})
}

//删除这个回调
func (c *Client) deleteWaitPackFunc(call *func(model.MsgID, []byte, net.Conn, error)) {

	value, ok := c.waitPackList.Load(call)
	if ok {
		v := value.(*WaitPackStr)
		go (*v.Call)(0, nil, nil, errors.New("cancel callback"))
		c.waitPackList.Delete(call)
	}

}

//发送消息
func (c *Client) Send(msgObj interface{}) error {
	b, err := codec.Encode(msgObj)
	if err != nil {
		return err
	}
	if !c.connected {
		return errors.New("client not connected")
	}
	_, err = c.sess.Write(b)
	return err
}

//关闭
func (c *Client) Connected() bool {
	return c.connected
}

func (c *Client) Close() {
	if !c.connected {
		return
	}
	c.tcpHandle(model.MsgID_DisconnectID, nil, c.sess)

	c.heartBeatTickerOver <- 0
	close(c.heartBeatTickerOver)

	c.waitPackTimeoutOver <- 0
	close(c.waitPackTimeoutOver)

	if c.sess != nil {
		_ = c.sess.Close()
	}
	c.sess = nil
	c.connected = false
}

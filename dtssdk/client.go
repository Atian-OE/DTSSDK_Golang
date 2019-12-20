package dtssdk

import (
	"bytes"
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/codec"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/utils"
	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/core/errors"
	"log"
	"net"
	"sync"
	"time"
)

func NewDTSClient(addr string) *Client {
	conn := &Client{
		port: 17083,
	}
	conn.init(addr)
	return conn
}

func (c *Client) init(addr string) {
	if c.Id() == "" {
		v4, err := uuid.NewV4()
		if err != nil {
			c.SetId("")
		} else {
			c.SetId(v4.String())
		}
	}
	c.addr = addr
	c.waitPackList = new(sync.Map)

	c.waitPackTimeoutTicker = time.NewTicker(time.Millisecond * 500)
	c.waitPackTimeoutOver = make(chan interface{})
	c.heartBeatTicker = time.NewTicker(time.Second * 5)
	c.heartBeatTickerOver = make(chan interface{})
	c.reconnectTicker = time.NewTicker(time.Second * c.ReconnectTime())
	c.reconnectTickerOver = make(chan interface{})

	go c.waitPackTimeout()
	go c.heartBeat()
	go c.reconnect()
}

func (c *Client) connect() {
	c.reconnecting = true
	if c.connected {
		return
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.addr, c.port), time.Second*3)
	if err != nil {
		if c.timeoutAction != nil {
			c.timeoutAction(c.addr)
		}
		return
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		if c.timeoutAction != nil {
			c.timeoutAction(c.addr)
		}
		return
	}
	c.sess = tcpConn
	//禁用缓存
	_ = tcpConn.SetWriteBuffer(5000)
	_ = tcpConn.SetReadBuffer(5000)
	go c.clientHandle(tcpConn)
}

func (c *Client) reconnect() {
	c.connected = false
	c.connect()
	for {
		select {
		case <-c.reconnectTicker.C:
			if !c.connected {
				c.count += 1
				if c.reconnectTimes == 0 {
					log.Println(fmt.Sprintf("[ 客户端%s ]正在无限尝试第[ %d/%d ]次重新连接[ %s ]...", c.Id(), c.count, c.reconnectTimes, c.addr))
					c.connect()
				} else {
					if c.count <= c.reconnectTimes {
						log.Println(fmt.Sprintf("[ 客户端%s ]正在尝试第[ %d/%d ]次重新连接[ %s ]...", c.Id(), c.count, c.reconnectTimes, c.addr))
						c.connect()
					} else {
						log.Println(fmt.Sprintf("[ 客户端%s ]第[ %d/%d ]次重新连接失败,断开连接[ %s ]...", c.Id(), c.count-1, c.reconnectTimes, c.addr))
						c.Close()
					}
				}
			}
		case <-c.reconnectTickerOver:
			log.Println(fmt.Sprintf("[ 客户端%s ]断开连接[ %s ]...", c.Id(), c.addr))
			return
		}
	}
}

//心跳
func (c *Client) heartBeat() {
	for {
		select {
		case <-c.heartBeatTicker.C:
			if c.connected {
				b, _ := codec.Encode(&model.HeartBeat{})
				_, _ = c.sess.Write(b)
			}

		case <-c.heartBeatTickerOver:
			return
		}
	}
}

//超时删除回调
func (c *Client) waitPackTimeout() {
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
			return

		}
	}
}

func (c *Client) clientHandle(conn net.Conn) {
	c.tcpHandle(model.MsgID_ConnectID, nil, conn)
	defer func() {
		if conn != nil {
			c.tcpHandle(model.MsgID_DisconnectID, nil, conn)
			_ = conn.Close()
		}
	}()

	buf := make([]byte, 1024)
	var cache bytes.Buffer
	for {
		//cache_index:=0
		n, err := conn.Read(buf)
		//加上上一次的缓存
		//n=buf_index+n
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
		return ErrClientNotConnect(c.addr)
	}
	_, err = c.sess.Write(b)
	return err
}

//关闭
func (c *Client) Close() {

	c.reconnectTicker.Stop()
	c.reconnectTickerOver <- 0
	close(c.reconnectTickerOver)

	c.heartBeatTicker.Stop()
	c.heartBeatTickerOver <- 0
	close(c.heartBeatTickerOver)

	c.waitPackTimeoutTicker.Stop()
	c.waitPackTimeoutOver <- 0
	close(c.waitPackTimeoutOver)

	if c.sess != nil {
		_ = c.sess.Close()
	}
	c.count = 0
	c.reconnecting = false
	c.connected = false
	c.sess = nil
}

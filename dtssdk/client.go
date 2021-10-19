package dtssdk

import (
	"bytes"
	"fmt"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/codec"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/model"
	"github.com/Atian-OE/DTSSDK_Golang/dtssdk/utils"
	"github.com/kataras/iris/core/errors"
	"net"
	"sync"
	"time"
)

type WaitPackStr struct {
	Key model.MsgID
	Timeout int64//毫秒
	Call *func(model.MsgID, []byte, net.Conn,error)
}



type DTSSDKClient struct{
	sess *net.TCPConn
	connected bool
	wait_pack_list *sync.Map//等待这个包回传
	wait_pack_timeout_ticker *time.Ticker//等待回传的回调 会在 3秒后 自动删除
	wait_pack_timeout_over chan interface{} //关闭自动删除
	heart_beat_ticker *time.Ticker//心跳包的发送
	heart_beat_ticker_over  chan interface{} //关闭心跳

	reconnect_ticker *time.Ticker//自动连接
	reconnect_ticker_over  chan interface{} //关闭自动连接



	addr string//地址


	_connected_action         func(string)                               //连接到服务器的回调
	_disconnected_action      func(string)                               //断开连接到服务器的回调
	_ZoneTempNotifyEnable    bool                                 //接收分区温度更新的通知
	_ZoneTempNotify          func(*model.ZoneTempNotify,error)    //分区温度更新
	_ZoneAlarmNotifyEnable   bool                                 //接收温度警报的通知
	_ZoneAlarmNotify         func(*model.ZoneAlarmNotify,error)   //分区警报通知
	_FiberStatusNotifyEnable bool                                 //接收设备状态改变的通知
	_FiberStatusNotify       func(*model.DeviceEventNotify,error) //设备状态通知
	_TempSignalNotifyEnable  bool                                 //接收设备温度信号的通知
	_TempSignalNotify        func(*model.TempSignalNotify,error)  //设备状态通知
	_ButtonNotifyEnable  bool                                 //接收设备温度信号的通知
	_ButtonNotify        func(*model.ButtonNotify,error)  //按钮通知
}

func NewDTSClient(addr string) *DTSSDKClient {
	conn:= &DTSSDKClient{}
	conn.init(addr)
	return conn
}

func(self *DTSSDKClient)init(addr string)  {
	self.addr=addr
	self.wait_pack_list=new(sync.Map)


	self.wait_pack_timeout_ticker= time.NewTicker(time.Millisecond*500)
	self.wait_pack_timeout_over=make(chan interface{})
	self.heart_beat_ticker= time.NewTicker(time.Second*5)
	self.heart_beat_ticker_over=make(chan interface{})
	self.reconnect_ticker=time.NewTicker(time.Second*10)
	self.reconnect_ticker_over=make(chan interface{})

	go self.wait_pack_timeout()
	go self.heart_beat()
	go self.reconnect()


}


func (self *DTSSDKClient)connect()  {
	if self.connected {
		return
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:17083",self.addr),time.Second*3)
	if err!=nil {
		//fmt.Println("连接服务器失败!")
		return
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		//fmt.Println("连接服务器失败!")
		return
	}
	self.sess=tcpConn
	//禁用缓存
	tcpConn.SetWriteBuffer(5000)
	tcpConn.SetReadBuffer(5000)
	go self.client_handle(tcpConn)
}

func (self *DTSSDKClient)reconnect()  {
	self.connected=false
	self.connect()

	for{
		select {
		case <-self.reconnect_ticker.C:
			self.connect()

		case <-self.reconnect_ticker_over:
			return

		}
	}
}

//心跳
func (self *DTSSDKClient)heart_beat() {
	for{
		select {
		case <-self.heart_beat_ticker.C:
			if(self.connected){
				b,_:=codec.Encode(&model.HeartBeat{})
				self.sess.Write(b)
			}

		case <-self.heart_beat_ticker_over:
			return

		}
	}
}

//超时删除回调
func (self *DTSSDKClient) wait_pack_timeout()  {
	for{
		select {
		case <-self.wait_pack_timeout_ticker.C:
			self.wait_pack_list.Range(func(key ,value interface{}) bool{

				v:=value.(*WaitPackStr)
				v.Timeout-=500
				if(v.Timeout<=0){
					go (*v.Call)(0,nil,nil,errors.New("callback timeout"))
					self.wait_pack_list.Delete(key)
				}
				return true
			})


		case <-self.wait_pack_timeout_over:
			return

		}
	}
}

func (self *DTSSDKClient)client_handle(conn net.Conn)  {
	self.tcp_handle(model.MsgID_ConnectID,nil,conn)
	defer func() {
		if(conn!=nil){
			self.tcp_handle(model.MsgID_DisconnectID,nil,conn)
			conn.Close()
		}
	}()


	buf:=make([]byte,1024)
	var cache bytes.Buffer
	for{
		//cache_index:=0
		n,err:=conn.Read(buf)
		//加上上一次的缓存
		//n=buf_index+n
		if(err!=nil){
			break
		}

		cache.Write(buf[:n])
		for{
			if(self.unpack(&cache,conn)){
				break
			}
		}
	}

}

// true 处理完成 false 循环继续处理
func (self *DTSSDKClient)unpack(cache *bytes.Buffer,conn net.Conn) bool {
	if(cache.Len()<5){
		return true
	}
	buf:=cache.Bytes()
	pkg_size:= utils.ByteToInt2(buf[:4])
	//长度不够
	if(pkg_size > len(buf)-5){
		return true
	}

	cmd:=buf[4]
	self.tcp_handle(model.MsgID(cmd),buf[:pkg_size+5],conn)
	cache.Reset()
	cache.Write(buf[5+pkg_size:])

	return false
}



//这个包会由这个回调接受
func (self*DTSSDKClient) wait_pack(msg_id model.MsgID,call *func(model.MsgID, []byte, net.Conn,error))  {
	self.wait_pack_list.Store(call,&WaitPackStr{Key:msg_id,Timeout:10000,Call:call})
}

//删除这个回调
func (self*DTSSDKClient) delete_wait_pack_func(call *func(model.MsgID, []byte, net.Conn,error)) ()  {

	value,ok:=self.wait_pack_list.Load(call)
	if(ok){
		v:=value.(*WaitPackStr)
		go (*v.Call)(0,nil,nil,errors.New("cancel callback"))
		self.wait_pack_list.Delete(call)
	}


}

//发送消息
func (self*DTSSDKClient)Send(msg_obj interface{}) error {
	b,err:=codec.Encode(msg_obj)
	if(err!=nil) {
		return err
	}
	if(!self.connected){
		return errors.New("client not connected")
	}
	_,err=self.sess.Write(b)
	return err
}


//关闭
func (self*DTSSDKClient)Close() {

	self.reconnect_ticker.Stop()
	self.reconnect_ticker_over<-0
	close(self.reconnect_ticker_over)

	self.heart_beat_ticker.Stop()
	self.heart_beat_ticker_over<-0
	close(self.heart_beat_ticker_over)

	self.wait_pack_timeout_ticker.Stop()
	self.wait_pack_timeout_over<-0
	close(self.wait_pack_timeout_over)

	if(self.sess!=nil){
		self.sess.Close()
	}

	self.sess=nil
}
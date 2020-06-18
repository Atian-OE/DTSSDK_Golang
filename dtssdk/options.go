package dtssdk

import "time"

type Options struct {
	Id          string
	Ip          string
	Port        int
	Timeout     time.Duration
	PingTime    time.Duration
	ReadBuffer  int
	WriteBuffer int
}

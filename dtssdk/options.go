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

func DefaultOptions(id, ip string) Options {
	return Options{
		Id:          id,
		Ip:          ip,
		Port:        17083,
		Timeout:     time.Second * 3,
		PingTime:    time.Minute,
		ReadBuffer:  2048,
		WriteBuffer: 2048,
	}
}

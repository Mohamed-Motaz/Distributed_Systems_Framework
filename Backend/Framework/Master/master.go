package main

import (
	mq "Framework/MessageQueue"
)

func NewMaster() *Master {
	master := &Master{
		q: mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
	}

	//start any go routines here

	return master
}

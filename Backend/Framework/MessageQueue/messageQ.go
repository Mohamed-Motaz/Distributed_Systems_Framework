package MessageQueue

import (
	logger "Server/Logger"
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func New(amqpAddr string) *MQ {
	mq := &MQ{
		conn: nil,
		ch:   nil,
		qMap: make(map[string]*amqp.Queue),
		mu:   sync.Mutex{},
	}

	//try to reconnect and sleep 2 seconds on failure
	var err error = fmt.Errorf("error")
	ctr := 0
	for ctr < 3 && err != nil {
		err = mq.connect(amqpAddr)
		ctr++
		if err != nil {
			time.Sleep(time.Second * 10)
		}
	}

	if err != nil {
		logger.FailOnError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to setup the MessageQueue %+v\n", amqpAddr)
	} else {
		logger.LogInfo(logger.MESSAGE_Q, logger.ESSENTIAL, "MessageQueue setup complete")
	}
	return mq
}

func (mq *MQ) Close() {
	mq.ch.Close()
	mq.conn.Close()
}

func (mq *MQ) Publish(qName string, body []byte) error {

	mq.mu.Lock()
	//if q is nill, declare it
	if _, ok := mq.qMap[qName]; !ok {
		mq.mu.Unlock()
		q, err := mq.ch.QueueDeclare(
			qName,
			true,  //durable
			false, //autoDelete
			false, //exclusive -> send errors when another consumer tries to connect to it
			false, //noWait
			nil,
		)
		if err != nil {
			return err
		}
		mq.mu.Lock()
		mq.qMap[qName] = &q
	}
	mq.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := mq.ch.PublishWithContext(
		ctx,
		"",    //exchange,   empty is default
		qName, //routing key
		false, //mandatory
		false, //immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	logger.LogInfo(logger.MESSAGE_Q, logger.NON_ESSENTIAL, "Successfully published to queue %v a message with this data:\n%v", qName, string(body))
	return nil
}

func (mq *MQ) connect(amqpAddr string) error {

	conn, err := amqp.Dial(amqpAddr)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Exiting because of err while establishing connection with message queue %v", err)
		return err
	}
	mq.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Exiting because of err while opening a channel %v", err)
		return err
	}
	mq.ch = ch

	logger.LogInfo(logger.MESSAGE_Q, logger.ESSENTIAL, "Connection to message queue has been successfully established")
	return nil
}

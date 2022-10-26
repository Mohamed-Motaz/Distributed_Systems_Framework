package MessageQueue

import (
	logger "Framework/Logger"
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewMQ(amqpAddr string) *MQ {
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

func CreateMQAddress(username string, password string, host string, port string) string {
	return "amqp://" + username + ":" + password + "@" + host + ":" + port + "/"
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

	logger.LogInfo(logger.MESSAGE_Q, logger.DEBUGGING, "Successfully published to queue %v a message with this data:\n%v", qName, string(body))
	return nil
}

func (mq *MQ) Consume(qName string) (<-chan amqp.Delivery, error) {

	//if q is nill, declare it
	mq.mu.Lock()
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
			return nil, err
		}
		mq.mu.Lock()
		mq.qMap[qName] = &q
	}
	mq.mu.Unlock()

	err := mq.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}
	msgs, err := mq.ch.Consume(
		qName,
		"",    //consumer --> unique identifier that rabbit can just generate
		false, //auto ack
		false, //exclusive --> errors if other consumers consume
		false, //nolocal, dont receive back if i send on channel
		false, //no wait
		nil,
	)
	if err != nil {
		return nil, err
	}

	logger.LogInfo(logger.MESSAGE_Q, logger.DEBUGGING, "Successfully subscribed to queue %v", qName)
	return msgs, nil
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
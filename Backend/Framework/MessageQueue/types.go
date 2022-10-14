package MessageQueue

import (
	utils "Server/Utils"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

//queue names
const (
	ENTITIES_QUEUE string = "entitiesQueue"
)

type DB_ACTION string

const (
	INSERT_ACTION DB_ACTION = "insert"
	UPDATE_ACTION DB_ACTION = "update"
)

type OBJ_TYPE string

const (
	CHAT         OBJ_TYPE = "chat"
	MESSAGE      OBJ_TYPE = "message"
	CHATS_CTR    OBJ_TYPE = "chatsCtr"
	MESSAGES_CTR OBJ_TYPE = "messageCtr"
)

type MQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	qMap map[string]*amqp.Queue
	mu   sync.Mutex
}

//object passed into the messageQ
type TransferObj struct {
	Action  DB_ACTION `json:"action"`
	ObjType OBJ_TYPE  `json:"objType"`
	Bytes   []byte    `json:"bytes"`
}

type Chat struct {
	Id                int       `json:"id"`
	Application_token string    `json:"application_token"`
	Number            int32     `json:"number"`
	Messages_count    int32     `json:"messages_count"`
	Created_at        time.Time `json:"created_at"`
	Updated_at        time.Time `json:"updated_at"`
}

type Message struct {
	Id         int       `json:"id"`
	Chat_id    int32     `json:"chat_id"`
	Number     int32     `json:"number"`
	Body       string    `json:"body"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type ChatsCtr struct {
	Token       string `json:"token"`
	Chats_count int32  `json:"chats_count"`
}

type MessagesCtr struct {
	Application_token string `gorm:"column:application_token"  json:"application_token"`
	Number            int32  `gorm:"column:number"             json:"number"`
	Messages_count    int32  `gorm:"column:messages_count"     json:"messages_count"`
}

const (
	MQ_PORT     string = "MQ_PORT"
	MQ_HOST     string = "MQ_HOST"
	LOCAL_HOST  string = "127.0.0.1"
	MQ_USERNAME string = "MQ_USERNAME"
	MQ_PASSWORD string = "MQ_PASSWORD"
)

var MqHost string
var MqPort string
var MqUsername string
var MqPassword string

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}
	MqHost = strings.Replace(utils.GetEnv(MQ_HOST, LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MqPort = utils.GetEnv(MQ_PORT, "5672")
	MqUsername = utils.GetEnv(MQ_USERNAME, "guest")
	MqPassword = utils.GetEnv(MQ_PASSWORD, "guest")

}

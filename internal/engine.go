package internal

import (
	"context"
	"fmt"
	"github.com/Dmitry-dms/avalanche-control/pkg/serializer"
	redis "github.com/go-redis/redis/v8"
	"log"
)

type Config struct {
	RedisAddress        string
	RedisInfoPrefix     string
	RedisMsgPrefix      string
	RedisCommandsPrefix string
}

type Engine struct {
	Context          context.Context
	Conf             Config
	Logger           *log.Logger
	Cache            *Cache
	MsgChannel       chan redisMessage
	RedisMsg         func(payload []byte) error
	RedisGetInfo     *redis.PubSub
	RedisAddCompany  func(company addCompanyResponse) error
	RedisCommandsSub func(payload []byte) error
	Serializer       serializer.AvalancheSerializer
}

func NewEngine(ctx context.Context, config Config, logger *log.Logger, cache *Cache) (*Engine, error) {
	red := initRedis(config.RedisAddress)

	redisMsg := func(payload []byte) error {
		return red.Publish(ctx, config.RedisMsgPrefix, payload).Err() //+config.Name
	}

	redisInfo := red.Subscribe(ctx, config.RedisInfoPrefix) //+config.Name)

	redisMain := func(payload []byte) error {
		return red.Publish(ctx, config.RedisCommandsPrefix, payload).Err() //+config.Name
	}

	if err := red.Ping(red.Context()).Err(); err != nil {
		logger.Println(err)
	}
	e := &Engine{
		Context:          ctx,
		Conf:             config,
		Logger:           logger,
		RedisMsg:         redisMsg,
		RedisGetInfo:     redisInfo,
		RedisCommandsSub: redisMain,
		Cache:            cache,
		MsgChannel:       make(chan redisMessage),
	}
	go e.sendMessages()
	go e.listeningCommands()
	logger.Println(fmt.Sprintf("Monitoring server succesfully connected to hub"))
	return e, nil
}

type redisMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
} //"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
type AddCompanyMessage struct {
	CompanyName string `json:"company_name"`
	MaxUsers    uint   `json:"max_users"`
	Duration    int    `json:"duration_hour"`
} //"{\"company_name\":\"testing\",\"max_users\":1000,\"duration_hour\":10}"
type CompanyToken struct {
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	Duration   int    `json:"duration_hour"`
}
type addCompanyResponse struct {
	Token       CompanyToken `json:"company_token"`
	CompanyName string       `json:"company_name"`
}

func (e *Engine) listeningCommands() {
	var info CompanyStats
	for msg := range e.RedisGetInfo.Channel() {
		err := e.Serializer.Deserialize([]byte(msg.Payload), &info)
		if err != nil {
			e.Logger.Println(err)
		}
	}
}
func (e *Engine) updateCache(stats CompanyStats) {
	e.Cache.update(&stats)
}
func (e *Engine) serializeAndSend(v interface{}, f func(payload []byte) error) error {
	payload, err := e.Serializer.Serialize(v)
	if err != nil {
		e.Logger.Println(err)
		return err
	}
	err = f(payload)
	if err != nil {
		e.Logger.Println(err)
		return err
	}
	return nil
}

func (e *Engine) sendMessages() {
	e.Logger.Println("listener was started")
	for msg := range e.MsgChannel {
		err := e.serializeAndSend(msg, e.RedisMsg)
		if err != nil {
			e.Logger.Println(err)
		}
		e.Logger.Printf("Message {%s} to client {%s} with company id {%s}", msg.Message, msg.ClientId, msg.CompanyName)
	}
}
func initRedis(address string) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return r
}

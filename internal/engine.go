package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Dmitry-dms/avalanche-control/pkg/serializer"
	redis "github.com/go-redis/redis/v8"
)

type Config struct {
	RedisAddress        string
	RedisInfoPrefix     string
	RedisMsgPrefix      string
	RedisCommandsPrefix string
}

type Engine struct {
	Context           context.Context
	Conf              Config
	Logger            *log.Logger
	Cache             *Cache
	MsgChannel        chan redisMessage
	AddCompanyChannel chan AddCompanyMessage
	RedisMsg          func(payload []byte) error
	RedisGetInfo      *redis.PubSub
	RedisAddCompany   func(company addCompanyResponse) error
	RedisCommands     func(payload []byte) error
	Serializer        serializer.AvalancheSerializer
}

func (e *Engine) Info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, e.Cache.Show())
}
func (e *Engine) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/msg", e.sendMessage)
	mux.HandleFunc("/add-company", e.addCompany)
	mux.HandleFunc("/", e.homePage)
	mux.HandleFunc("/i", e.Info)
	return mux
}



func (e *Engine) sendMessage(w http.ResponseWriter, r *http.Request) {
	var message redisMessage
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&message)
	if err != nil {
		printError(w, "Can't decode json", 400) // 400 Bad Request
		return
	}
	company, user := e.Cache.GetUser(message.CompanyName, message.ClientId)
	if !company {
		printError(w, "Company doesn't exists", 404) // 404 Not Found
		return
	}
	if !user {
		printError(w, "User doesn't exists", 404) // 404 Not Found
		return
	}
	e.MsgChannel <- redisMessage{
		CompanyName: message.CompanyName,
		ClientId: message.ClientId,
		Message: message.Message,
	}
	w.Write([]byte("Success"))
}
func (e *Engine) addCompany(w http.ResponseWriter, r *http.Request) {
	var message AddCompanyMessage
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&message)
	if err != nil {
		printError(w, "Can't decode json", 400) // 400 Bad Request
		return
	}
	company := e.Cache.GetCompany(message.CompanyName)
	if company {
		printError(w, "Company already exists", 400) // 400 Bad Request
		return
	}
	e.AddCompanyChannel <- AddCompanyMessage{
		CompanyName: message.CompanyName,
		MaxUsers: message.MaxUsers,
		Duration: message.Duration,
	}
	w.Write([]byte("Information was send. Please wait a moment"))
}
func printError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	str := fmt.Sprintf(`{"Reason":"%s"}`, msg)
	fmt.Fprint(w, str)
}
func (e *Engine) homePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home Page"))
}

func NewEngine(ctx context.Context, config Config, logger *log.Logger, cache *Cache, ser serializer.AvalancheSerializer) (*Engine, error) {
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
		Context:           ctx,
		Conf:              config,
		Logger:            logger,
		RedisMsg:          redisMsg,
		RedisGetInfo:      redisInfo,
		RedisCommands:     redisMain,
		Cache:             cache,
		MsgChannel:        make(chan redisMessage),
		AddCompanyChannel: make(chan AddCompanyMessage),
		Serializer:        ser,
	}
	go e.sendMessages()
	go e.getInfo()
	go e.sendCommands()
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

func (e *Engine) getInfo() {
	var info []CompanyStats
	var addC addCompanyResponse
	for msg := range e.RedisGetInfo.Channel() {
		go func(msg *redis.Message) {
			err := e.Serializer.Deserialize([]byte(msg.Payload), &info)
			if err != nil {
				err2 := e.Serializer.Deserialize([]byte(msg.Payload), &addC)
				if err2 != nil {
					e.Logger.Println(err2)
				}
				e.Logger.Printf("Added {%s} with token = %s, duration = %d, ws server: %s", addC.CompanyName, addC.Token.Token, addC.Token.Duration, addC.Token.ServerName)
			}
			for i := range info {
				e.updateCache(info[i].Name, &info[i])
			}
		}(msg)
	}
}
func (e *Engine) updateCache(companyName string, stats *CompanyStats) {
	e.Cache.update(companyName, stats)
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
func (e *Engine) sendCommands() {
	for msg := range e.AddCompanyChannel {
		err := e.serializeAndSend(msg, e.RedisCommands)
		if err != nil {
			e.Logger.Println(err)
		}
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

package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Dmitry-dms/avalanche-control/internal"
	"github.com/Dmitry-dms/avalanche-control/pkg/serializer/json"
)

func main() {
	c := internal.NewCache()
	ctx := context.Background()
	conf := internal.Config{
		RedisAddress:        "localhost:6050",
		RedisInfoPrefix:     "info",
		RedisMsgPrefix:      "msg",
		RedisCommandsPrefix: "comm",
	}
	ser := &json.CustomJsonSerializer{}
	log := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	eng, err := internal.NewEngine(ctx, conf, log, c, ser)
	if err != nil {
		log.Fatal(err)
	}
	mux := eng.NewServeMux()
	log.Println(http.ListenAndServe(":8040", mux))
}


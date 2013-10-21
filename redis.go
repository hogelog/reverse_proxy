package main

import (
	"log"
	"os"
//	"net/http"
//	"net/http/httputil"
//	"net/url"
	"github.com/garyburd/redigo/redis"
)

func dead(v ...interface{}) {
	log.Fatal(v)
	os.Exit(1)
}

func connectRedis(network, address string) {
	conn, err := redis.Dial(network, address)
	if err != nil {
		dead(err)
	}
	sessions, err := redis.Strings(conn.Do("SMEMBERS", "login_session"))
	if err != nil {
		dead(err)
	}
	for _, session := range sessions {
		log.Printf("session = %v", session)
	}
	defer conn.Close()
}

func main() {
	connectRedis("tcp", "localhost:6379")
}

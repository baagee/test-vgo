package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	Storage Storage
}

//从环境变量获取redis配置 并且获取redis链接
func getEnv() *Env {
	addr := os.Getenv("REDIS_ADDRESS")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = ""
	}
	dbS := os.Getenv("REDIS_DB")
	if dbS == "" {
		dbS = "2"
	}
	db, err := strconv.Atoi(dbS)
	if err != nil {
		panic("redis db not integer")
	}
	log.Printf("redis config address:%s password:%s db:%d\n", addr, password, db)
	redis := NewRedisCli(addr, password, db)
	return &Env{Storage: redis}
}

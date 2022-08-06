package main

import (
	"context"
	"fmt"
	"github.com/ciffelia/logport/agent/docker/log"
	"github.com/docker/docker/client"
	"time"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	stream, err := log.NewStream(cli, context.Background(), "with-tty")
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		if err := stream.Close(); err != nil {
			panic(err)
		}
	}()

	for stream.Wait() {
		line := stream.Get()
		fmt.Println(line.Timestamp, line.Payload)
	}

	if err := stream.Err(); err != nil {
		panic(err)
	}
}

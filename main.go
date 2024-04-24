package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// init websocket connect
	serverAddr := os.Getenv("SERVER")
	if len(serverAddr) == 0 {
		log.Fatalln("Env SERVER not found!")
	}
	log.Printf("start connecting to %s\n", serverAddr)
	c, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	agentTag := os.Getenv("TAG")
	if len(agentTag) == 0 {
		log.Fatalln("Env TAG not found!")
	}
	log.Printf("agent tag is %s\n", agentTag)

	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := c.WriteMessage(websocket.TextMessage, []byte(agentTag)); err != nil {
			log.Fatal("Register client error:", err)
		}
		var recvCMD ReceiveCMD
		for {
			err := c.ReadJSON(&recvCMD)
			if err != nil {
				log.Println("read msg error:", err)
				return
			}
			log.Printf("recv msg success: %+v", recvCMD)

			err = operateCMD(&recvCMD)
			if err != nil {
				c.WriteJSON(ResultCMD{
					Code: ResultFailure,
					Msg:  err.Error(),
				})
				continue
			}
			c.WriteJSON(ResultCMD{
				Code: ResultSuccess,
				Msg:  "",
			})
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// heartbeat
			err := c.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				log.Println("heartbeat error:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

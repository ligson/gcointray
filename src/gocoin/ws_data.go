package gocoin

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func ReadWss() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "ws.coincap.io", Path: "/prices", RawQuery: "assets=ALL"}
	log.Printf("connecting to %s", u.String())

	println(u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			data := make(map[string]string)
			// 将字节数组解析为 map
			err = json.Unmarshal(message, &data)
			if err != nil {
				fmt.Println("解析错误:", err)
				return
			}

			// 打印键和值
			for key, value := range data {
				goCoin, err := QueryCoinById(key)
				if err != nil {
					log.Fatal("查询" + key + "失败:" + err.Error())
				} else {
					log.Println("数字货币" + goCoin.Symbol + "价格是:" + value)
					updateCoinPrice(goCoin.Symbol, value)
				}
			}

		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupted")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close error:", err)
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

package main

import (
	"fmt"
	"github.com/getlantern/systray/example/icon"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

func main() {
	go readWss()
	// 初始化systray
	systray.Run(onReady, onExit)

}

func readWss() {
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
			log.Printf("received: %s", message)
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

func onReady() {
	println("ok.....")
	// 在任务栏托盘中创建图标
	systray.SetIcon(icon.Data)
	systray.SetTitle("myapp") // 设置图标标题
	time.Sleep(time.Second)
	systray.SetTooltip("比特币价格") // 设置鼠标悬停提示

	var coins []string = []string{"bitcoin", "ethereum", "solana", "terra-virtua-kolect"}
	var coinItems []*systray.MenuItem = []*systray.MenuItem{}
	for _, coin := range coins {
		// 创建一个菜单项，用于显示文字
		item := systray.AddMenuItem(coin+"---[0]", coin)
		coinItems = append(coinItems, item)
	}

	// 创建一个子菜单，用于点击弹出下拉列表
	subMenu := systray.AddMenuItem("Select", "")
	subMenu1 := subMenu.AddSubMenuItem("Option 1", "Option 1")
	subMenu.AddSubMenuItem("Option 2", "Option 2")

	// 创建右键菜单，显示另外下拉列表
	rightClickMenu := systray.AddMenuItem("Right Click", "")
	rightClickMenu.AddSubMenuItem("Option A", "Option A")
	rightClickMenu.AddSubMenuItem("Option B", "Option B")

	quitMenu := systray.AddMenuItem("退出", "退出应用")

	for idx, item := range coinItems {

		go func(coinItem *systray.MenuItem, index int) {
			<-coinItem.ClickedCh
			println(coins[index])
		}(item, idx)
	}
	// 在另外的协程中处理菜单项的点击事件
	go func() {
		for {
			select {
			case <-subMenu.ClickedCh:
				// 处理下拉列表的点击事件
				fmt.Println("Sub menu clicked")
			case <-subMenu1.ClickedCh:
				fmt.Println("Sub menu clicked 111")
			case <-quitMenu.ClickedCh:
				systray.Quit()
			case <-rightClickMenu.ClickedCh:
				// 处理右键菜单的点击事件
				fmt.Println("Right click menu clicked")
			}

		}
	}()
}

func onExit() {
	println("exit.....")
	// 清理工作并退出应用
	os.Exit(0)
}

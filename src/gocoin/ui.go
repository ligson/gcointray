package gocoin

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"os"
	"time"
)

var coins []string = []string{"BTC", "ETH", "SOL"}
var coinItems []*systray.MenuItem = []*systray.MenuItem{}

func updateCoinPrice(symbol string, price string) {
	for idx, item := range coinItems {
		gocoin := coins[idx]
		if gocoin == symbol {
			item.SetTitle(gocoin + "---" + price)
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

	for _, coin := range coins {
		// 创建一个菜单项，用于显示文字
		gocoins, err := QueryCoinsBySymbol(coin)
		itemTitle := ""
		if err != nil {
			itemTitle = coin + "---0"
		} else {
			if len(gocoins) > 0 {
				itemTitle = coin + "---" + gocoins[0].PriceUsd
			} else {
				itemTitle = coin + "---0"
			}
		}
		item := systray.AddMenuItem(itemTitle, coin)
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

func StartUI() {
	systray.Run(onReady, onExit)
}

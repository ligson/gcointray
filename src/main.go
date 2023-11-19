package main

import (
	"encoding/json"
	"fmt"
	"gcointray/src/db"
	"gcointray/src/model"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func loadCoin(offset int) (model.AssetsResponse, error) {
	// 解析JSON数据到结构体
	var post model.AssetsResponse
	reqUrl := "https://api.coincap.io/v2/assets?limit=2000&offset=" + strconv.Itoa(offset)
	log.Println("请求地址:" + reqUrl)
	// 发送HTTP GET请求
	resp, err := http.Get(reqUrl)
	if err != nil {
		log.Fatal(err)
		return post, err
	}
	defer resp.Body.Close()

	// 读取响应的JSON数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return post, err
	}

	log.Println(string(body))
	// 解析JSON数据到结构体
	err = json.Unmarshal(body, &post)
	if err != nil {
		log.Fatal(err)
		return post, err
	}
	return post, nil
}
func loadCoinDatas() []model.Coin {
	var coins []model.Coin
	offset := 1
	for {
		response, err := loadCoin(offset)
		if err != nil {
			return []model.Coin{}
		} else {
			data := response.Data
			coins = append(coins, data...)
			if len(data) == 2000 {
				offset = offset + 2000
			} else {
				break
			}
		}
	}
	return coins
}

func getAllCoin() {
	datas := loadCoinDatas()
	log.Println(len(datas))
	db, err := db.NewSQLiteDB("coin.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	// 创建表 t_coin
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS t_coin (
			id VARCHAR(100) PRIMARY KEY,
			rank VARCHAR(100),
			symbol VARCHAR(100),
			name VARCHAR(100),
			supply VARCHAR(100),
			max_supply VARCHAR(100),
			market_cap_usd VARCHAR(100),
			volume_usd_24hr VARCHAR(100),
			price_usd VARCHAR(100),
			change_percent_24hr VARCHAR(100),
			vwap_24hr VARCHAR(100),
			explorer VARCHAR(100)
		);
	`)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 查询t_coin表中的所有id
	rows, err := db.Query("SELECT id FROM t_coin")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// 创建一个map用于存储已存在的id
	existingIds := make(map[string]bool)
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		existingIds[id] = true
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	// 创建一个数组，用于存储需要插入的coins
	var coinsToInsert []model.Coin

	// 遍历Coin结构体数组，将不在existingIds中的coin添加到coinsToInsert中
	for _, coin := range datas {
		if !existingIds[coin.Id] {
			coinsToInsert = append(coinsToInsert, coin)
		}
	}

	// 构建批量插入的SQL语句
	var valueStrings []string
	var valueArgs []any

	for _, coin := range coinsToInsert {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, coin.Id, coin.Rank, coin.Symbol, coin.Name, coin.Supply, coin.MaxSupply,
			coin.MarketCapUsd, coin.VolumeUsd24Hr, coin.PriceUsd, coin.ChangePercent24Hr,
			coin.Vwap24Hr, coin.Explorer)
	}

	// 执行批量插入
	if len(coinsToInsert) > 0 {
		sqlQuery := "INSERT INTO t_coin (id, rank, symbol, name, supply, max_supply, market_cap_usd, volume_usd_24hr, price_usd, change_percent_24hr, vwap_24hr, explorer) VALUES " + strings.Join(valueStrings, ", ")
		_, err = db.Exec(sqlQuery, valueArgs...)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d coins inserted into t_coin table.", len(coinsToInsert))
	} else {
		log.Println("No coins to insert into t_coin table.")
	}

}

func main() {
	getAllCoin()
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

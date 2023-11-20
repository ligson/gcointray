package gocoin

import (
	"encoding/json"
	"gcointray/src/db"
	"gcointray/src/model"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// 解析JSON数据到结构体
	err = json.Unmarshal(body, &post)
	if err != nil {
		log.Fatal(err)
		return post, err
	}
	return post, nil
}
func loadCoinDatas() []model.GoCoin {
	var coins []model.GoCoin
	offset := 0
	for {
		response, err := loadCoin(offset)
		if err != nil {
			return []model.GoCoin{}
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

func getGoCoinDB() (*db.SQLiteDB, error) {
	return db.NewSQLiteDB("gocoin.db")
}
func UpdateCoins() {
	datas := loadCoinDatas()
	log.Println(len(datas))
	sqLiteDB, err := getGoCoinDB()
	if err != nil {
		log.Fatal(err)
		return
	}
	// 创建表 t_coin
	_, err = sqLiteDB.Exec(`
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
	rows, err := sqLiteDB.Query("SELECT id FROM t_coin")
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
	var coinsToInsert []model.GoCoin

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
		_, err = sqLiteDB.Exec(sqlQuery, valueArgs...)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d coins inserted into t_coin table.", len(coinsToInsert))
	} else {
		log.Println("No coins to insert into t_coin table.")
	}

}

var coinMap = make(map[string]model.GoCoin)

func QueryCoinById(coinId string) (model.GoCoin, error) {
	value, exists := coinMap[coinId]
	if exists {
		// 键存在于 map 中
		return value, nil
	}

	var coin model.GoCoin
	sqLiteDB, err := getGoCoinDB()
	if err != nil {
		log.Fatal(err)
		return model.GoCoin{}, err
	}
	rows, err := sqLiteDB.Query("select id, rank, symbol, name, supply, max_supply, market_cap_usd, volume_usd_24hr, price_usd, change_percent_24hr, vwap_24hr, explorer from t_coin where id=?", coinId)
	if err != nil {
		log.Fatal(err)
		return model.GoCoin{}, err
	}

	for rows.Next() {
		err = rows.Scan(
			&coin.Id,
			&coin.Rank,
			&coin.Symbol,
			&coin.Name,
			&coin.Supply,
			&coin.MaxSupply,
			&coin.MarketCapUsd,
			&coin.VolumeUsd24Hr,
			&coin.PriceUsd,
			&coin.ChangePercent24Hr,
			&coin.Vwap24Hr,
			&coin.Explorer,
		)
		if err != nil {
			defer sqLiteDB.Close()
			return model.GoCoin{}, err
		} else {
			break
		}
	}

	coinMap[coinId] = coin
	return coin, nil
}

var coinSymbolMap = make(map[string][]model.GoCoin)

func QueryCoinsBySymbol(symbol string) ([]model.GoCoin, error) {
	value, exists := coinSymbolMap[symbol]
	if exists {
		// 键存在于 map 中
		return value, nil
	}

	sqLiteDB, err := getGoCoinDB()
	if err != nil {
		log.Fatal(err)
		return []model.GoCoin{}, err
	}
	rows, err := sqLiteDB.Query("select id, rank, symbol, name, supply, max_supply, market_cap_usd, volume_usd_24hr, price_usd, change_percent_24hr, vwap_24hr, explorer from t_coin where symbol=?", symbol)
	if err != nil {
		log.Fatal(err)
		return []model.GoCoin{}, err
	}
	var goCoins []model.GoCoin = []model.GoCoin{}
	for rows.Next() {
		var coin model.GoCoin
		err = rows.Scan(
			&coin.Id,
			&coin.Rank,
			&coin.Symbol,
			&coin.Name,
			&coin.Supply,
			&coin.MaxSupply,
			&coin.MarketCapUsd,
			&coin.VolumeUsd24Hr,
			&coin.PriceUsd,
			&coin.ChangePercent24Hr,
			&coin.Vwap24Hr,
			&coin.Explorer,
		)
		if err != nil {
			return []model.GoCoin{}, err
		} else {
			goCoins = append(goCoins, coin)
		}
	}
	coinSymbolMap[symbol] = goCoins
	return goCoins, nil
}

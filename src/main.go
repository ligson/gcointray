package main

import (
	"gcointray/src/gocoin"
)

func main() {
	gocoin.UpdateCoins()
	go gocoin.ReadWss()
	// 初始化systray
	gocoin.StartUI()

}

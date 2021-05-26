package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

func (c *client) read() {
	// socketのメッセージ読み取りを開始する
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			// roomのメッセージ受信したらroom構造体のチャネルへ流す
			c.room.forward <- msg
			log.Println(msg)
		} else {
			// errの場合
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			// メッセージの送信にエラーがあれば中断
			break
		}
	}
	c.socket.Close()
}

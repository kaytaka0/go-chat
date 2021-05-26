package main

import (
	"chatapp/trace"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	forward chan []byte
	join    chan *client // roomにこれから参加するclientオブジェクト
	leave   chan *client // roomから退出するclientオブジェクト
	clients map[*client]bool
	tracer  trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// ルームに参加
			r.clients[client] = true
			r.tracer.Trace(("新しいクライアントが参加しました"))
		case client := <-r.leave:
			delete(r.clients, client) // 退出するのでroom.clientsから削除
			close(client.send)
			r.tracer.Trace(("クライアントが退室しました"))
		case msg := <-r.forward:
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信する
					r.tracer.Trace(("-- クライアントに送信されました"))
				default:
					// 送信に失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// HTTP接続をアップグレードしてwebsocket接続を取得する
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	// clientオブジェクトの作成
	client := &client{socket: socket, send: make(chan []byte, messageBufferSize), room: r}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

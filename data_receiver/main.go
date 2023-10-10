package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/atillatahak/tolling/types"
	"github.com/gorilla/websocket"
)

func main() {
	recv := NewDataReceiver()
	http.HandleFunc("/ws", recv.handleWS)
	http.ListenAndServe(":30000", nil)
}

type DataReceiver struct {
	msgch chan types.OBUData
	conn *websocket.Conn
}

func NewDataReceiver() *DataReceiver {
	return &DataReceiver{
		msgch: make(chan types.OBUData,128),
	}
}
func (dr *DataReceiver) handleWS(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	dr.conn = conn

	go dr.wsReceiverLoop()
}

func(dr *DataReceiver) wsReceiverLoop() {
	fmt.Println("wsReceiverLoop started")
	for {
		data := types.OBUData{}
		if err := dr.conn.ReadJSON(&data); err != nil {
			log.Fatal(err)
			continue
		}
		fmt.Printf("received data: [%d]\n :: <lat %.2f, long %.2f>", data.OBUID, data.Lat, data.Long)
		dr.msgch <- data
	}
}

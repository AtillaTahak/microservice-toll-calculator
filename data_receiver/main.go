package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atillatahak/tolling/types"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)
const kafkaTopic = "obudata"
const partition = 0

func main() {

	recv := NewDataReceiver()
	http.HandleFunc("/ws", recv.handleWS)
	http.ListenAndServe(":30000", nil)
}

type DataReceiver struct {
	msgch chan types.OBUData
	conn *websocket.Conn
	prod *kafka.Conn
}

func NewDataReceiver() *DataReceiver {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", kafkaTopic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	return &DataReceiver{
		msgch: make(chan types.OBUData,128),
		prod: conn,
	}
}
func (dr *DataReceiver) produceData(data types.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	
	dr.prod.SetWriteDeadline(time.Now().Add(10*time.Second))
	_, err = dr.prod.WriteMessages(
		kafka.Message{
			Value: b,
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	return nil
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
		if err := dr.produceData(data); err != nil {
			fmt.Println("failed to produce data:", err)
		}
		fmt.Printf("received data: [%d]\n :: <lat %.2f, long %.2f>", data.OBUID, data.Lat, data.Long)
		//dr.msgch <- data
	}
}

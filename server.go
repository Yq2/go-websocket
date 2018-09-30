package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"go-websocket/impl"
	"time"
)

//定义转换器
var (
	upgrader = websocket.Upgrader {
		//允许跨域,从直播.com 跨域到 websocket.com
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)


func wsHandler (w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("hello yangqiang"))
	//升级http协议到websocket协议
	var (
		wsConn *websocket.Conn
		err error
		//msgType int
		data []byte
		conn *impl.Connection
	)
	//websocketConn
	if wsConn, err = upgrader.Upgrade(w,r,nil); err != nil {
		return
	}
	if conn ,err = impl.InitConnection(wsConn); err != nil {
		goto WSERR
	}
	//心跳
	go func() {
		var (
			err error
		)
		for {
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		if data ,err = conn.ReadMessage(); err != nil {
			goto WSERR
		}
		if err = conn.WriteMessage(data); err != nil {
			goto WSERR
		}
	}

	WSERR:
		// TODO 关闭连接的操作
		conn.Close()

	//for {
	//	//Text Binary
	//	if _, data, err = conn.ReadMessage(); err != nil {
	//		goto ERR
	//	}
	//	if err = conn.WriteMessage(websocket.TextMessage,data); err != nil {
	//		goto ERR
	//	}
	//}
	//ERR:
	//	conn.Close()

}

func main () {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("0.0.0.0:7777", nil)
}

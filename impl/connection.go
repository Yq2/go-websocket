package impl

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

//channel 是线程安全的

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte //二进制消息
	outChan   chan []byte
	closeChan chan byte
	isClosed  bool
	mutex     sync.Mutex
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000), //1000个消息的容量
		outChan:   make(chan []byte, 1000), //
		closeChan: make(chan byte, 1),
	}
	//启动协程
	go conn.readLoop()
	//启动协程
	go conn.writeLoop()
	return
}

//线程安全的
func (conn *Connection) ReadMessage() (data []byte, err error) {
	//data = <- conn.inChan
	//改进  防止底层连接错误 用户阻塞在这里
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed!")
	}
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	//conn.outChan <- data
	//改进  防止底层连接错误 用户阻塞在这里
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed!")
	}
	return
}

func (conn *Connection) Close() {
	//线程安全 可以重入
	conn.wsConn.Close() // 这个close是线程安全的
	//close(conn.closeChan)
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

//内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		//阻塞在这里，等待取走
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: //当closeChan 被关闭时
			goto ERR
		}
		//conn.inChan <- data
	}
ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
	)
	for {
		//data  = <- conn.outChan
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		if err := conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
package msggateway

import (
	"io"
	"net"
	"time"
)

type TCPConnWarp struct {
	conn *net.TCPConn
}

func (t *TCPConnWarp) Close() error {
	return t.conn.Close()
}

func (t *TCPConnWarp) WriteMessage(_ int, message []byte) error {
	_, err := t.conn.Write(message)
	return err
}

func (t *TCPConnWarp) ReadMessage() (int, []byte, error) {
	data, err := io.ReadAll(t.conn)
	return MessageBinary, data, err
}

func (t *TCPConnWarp) SetReadDeadline(timeout time.Duration) error {
	return t.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (t *TCPConnWarp) SetWriteDeadline(timeout time.Duration) error {
	return t.conn.SetWriteDeadline(time.Now().Add(timeout))
}

func (t *TCPConnWarp) IsNil() bool {
	return t.conn == nil
}

func (t *TCPConnWarp) SetConnNil() {
	t.conn = nil
}

func (t *TCPConnWarp) SetReadLimit(limit int64) {
}

func (t *TCPConnWarp) SetPongHandler(handler PingPongHandler) {

}

func (t *TCPConnWarp) SetPingHandler(handler PingPongHandler) {
}

type TcpServer struct {
	addr        string
	clientAgent ConnClientAgent
	ln          net.Listener
}

func NewTCPServer(addr string, clientAgent ConnClientAgent) *TcpServer {
	return &TcpServer{addr: addr, clientAgent: clientAgent}
}

func (t *TcpServer) Run() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", t.addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	t.ln = ln
	agent := t.clientAgent
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			// "Accept failed: ", err
			continue
		}

		lconn := &TCPConnWarp{conn: conn}
		// todo hand build for user context
		agent.CreateNewConnClient(nil, lconn)
		//news := newSession(conn, serverInfo.CloseChan, kv)
		//// "New session: %v %v.", news.Id, news.ip
		//serverInfo.ConnectChan <- news
	}

	return nil
}

func (t *TcpServer) Close() error {
	return t.ln.Close()
}

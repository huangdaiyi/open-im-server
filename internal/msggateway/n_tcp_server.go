package msggateway

import (
	"net"
)

type TCPConn struct {
	Listener *net.TCPListener
}

func NewTCPServer(addr string) (*TCPConn, error) {
	serverInfo := TCPConn{}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	serverInfo.Listener = listener

	go func() {
		for {
			_, err := listener.AcceptTCP()
			if err != nil {
				// "Accept failed: ", err
				continue
			}
			//news := newSession(conn, serverInfo.CloseChan, kv)
			//// "New session: %v %v.", news.Id, news.ip
			//serverInfo.ConnectChan <- news
		}
	}()

	return &serverInfo, nil
}

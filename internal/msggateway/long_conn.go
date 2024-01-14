// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package msggateway

import (
	"encoding/json"
	"fmt"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/apiresp"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type LongConn interface {
	// Close this connection
	Close() error
	// WriteMessage Write message to connection,messageType means data type,can be set binary(2) and text(1).
	WriteMessage(messageType int, message []byte) error
	// ReadMessage Read message from connection.
	ReadMessage() (int, []byte, error)
	// SetReadDeadline sets the read deadline on the underlying network connection,
	// after a read has timed out, will return an error.
	SetReadDeadline(timeout time.Duration) error
	// SetWriteDeadline sets to write deadline when send message,when read has timed out,will return error.
	SetWriteDeadline(timeout time.Duration) error
	// IsNil Whether the connection of the current long connection is nil
	IsNil() bool
	// SetConnNil Set the connection of the current long connection to nil
	SetConnNil()
	// SetReadLimit sets the maximum size for a message read from the peer.bytes
	SetReadLimit(limit int64)
	SetPongHandler(handler PingPongHandler)
	SetPingHandler(handler PingPongHandler)
	// doUpgrade Check the connection of the current and when it was sent are the same
	//doUpgrade(w http.ResponseWriter, r *http.Request) error
}

type GWebSocket struct {
	handshakeTimeout time.Duration
	writeBufferSize  int
	cache            cache.MsgModel
	newClient        func(ctx *UserConnContext)
}

func newGWebSocket(handshakeTimeout time.Duration, wbs int) *GWebSocket {
	return &GWebSocket{handshakeTimeout: handshakeTimeout, writeBufferSize: wbs}
}

func (d *GWebSocket) doUpgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader := &websocket.Upgrader{
		HandshakeTimeout: d.handshakeTimeout,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	if d.writeBufferSize > 0 { // default is 4kb.
		upgrader.WriteBufferSize = d.writeBufferSize
	}

	return upgrader.Upgrade(w, r, nil)
}

func (d *GWebSocket) wsHandler(w http.ResponseWriter, r *http.Request) {
	args, pErr := d.ParseWSArgs(r)
	if pErr != nil {
		return
	}

	connContext := newContext(args.UserID, r.RemoteAddr)
	conn, err := d.doUpgrade(w, r)
	if err != nil {
		httpError(connContext, err)
		return
	}

	if args.MsgResp {
		data, err := json.Marshal(apiresp.ParseError(pErr))
		if err != nil {
			_ = conn.Close()
			return
		}
		if err := conn.WriteMessage(MessageText, data); err != nil {
			_ = conn.Close()
			return
		}
	} else {
		if pErr != nil {
			httpError(connContext, pErr)
			return
		}
		//wsLongConn = newGWebSocket(WebSocket, ws.handshakeTimeout, ws.writeBufferSize)
		//if err := wsLongConn.doUpgrade(w, r); err != nil {
		//	httpError(connContext, err)
		//	return
		//}
	}

}

func (d *GWebSocket) ParseWSArgs(r *http.Request) (args *WSArgs, err error) {
	var v WSArgs
	defer func() {
		args = &v
	}()
	query := r.URL.Query()
	v.MsgResp, _ = strconv.ParseBool(query.Get(MsgResp))
	//if ws.onlineUserConnNum.Load() >= ws.maxConnNum {
	//	return nil, errs.ErrConnOverMaxNumLimit.Wrap("over max conn num limit")
	//}
	if v.Token = query.Get(Token); v.Token == "" {
		return nil, errs.ErrConnArgsErr.Wrap("token is empty")
	}
	if v.UserID = query.Get(WsUserID); v.UserID == "" {
		return nil, errs.ErrConnArgsErr.Wrap("sendID is empty")
	}
	platformIDStr := query.Get(PlatformID)
	if platformIDStr == "" {
		return nil, errs.ErrConnArgsErr.Wrap("platformID is empty")
	}
	platformID, err := strconv.Atoi(platformIDStr)
	if err != nil {
		return nil, errs.ErrConnArgsErr.Wrap("platformID is not int")
	}
	v.PlatformID = platformID
	if err = authverify.WsVerifyToken(v.Token, v.UserID, platformID); err != nil {
		return nil, err
	}
	if query.Get(Compression) == GzipCompressionProtocol {
		v.Compression = true
	}
	if r.Header.Get(Compression) == GzipCompressionProtocol {
		v.Compression = true
	}
	m, err := ws.cache.GetTokensWithoutError(context.Background(), v.UserID, platformID)
	if err != nil {
		return nil, err
	}
	if v, ok := m[v.Token]; ok {
		switch v {
		case constant.NormalToken:
		case constant.KickedToken:
			return nil, errs.ErrTokenKicked.Wrap()
		default:
			return nil, errs.ErrTokenUnknown.Wrap(fmt.Sprintf("token status is %d", v))
		}
	} else {
		return nil, errs.ErrTokenNotExist.Wrap()
	}
	return &v, nil
}

type WSArgs struct {
	Token       string
	UserID      string
	PlatformID  int
	Compression bool
	MsgResp     bool
}

// WsConn websocket 连接。
type WsConn struct {
	conn *websocket.Conn
}

func (w *WsConn) Close() error {
	return w.conn.Close()
}

func (w *WsConn) WriteMessage(messageType int, message []byte) error {
	return w.conn.WriteMessage(messageType, message)
}

func (w *WsConn) ReadMessage() (int, []byte, error) {
	return w.conn.ReadMessage()
}

func (w *WsConn) SetReadDeadline(timeout time.Duration) error {
	return w.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (w *WsConn) SetWriteDeadline(timeout time.Duration) error {
	return w.conn.SetWriteDeadline(time.Now().Add(timeout))
}

func (w *WsConn) IsNil() bool {
	return w.conn == nil
}

func (w *WsConn) SetConnNil() {
	w.conn = nil
}

func (w *WsConn) SetReadLimit(limit int64) {
	w.conn.SetReadLimit(limit)
}

func (w *WsConn) SetPongHandler(handler PingPongHandler) {
	w.conn.SetPongHandler(handler)
}

func (w *WsConn) SetPingHandler(handler PingPongHandler) {
	w.conn.SetPingHandler(handler)
}

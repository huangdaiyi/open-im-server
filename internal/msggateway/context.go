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
	"strconv"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/utils"
)

type UserConnContext struct {
	//RespWriter http.ResponseWriter
	//Req        *http.Request
	platformID   string
	operationID  string
	userID       string
	token        string
	RemoteAddr   string
	ConnID       string
	isBackground bool
	isCompress   bool
}

func (c *UserConnContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *UserConnContext) Done() <-chan struct{} {
	return nil
}

func (c *UserConnContext) Err() error {
	return nil
}

func (c *UserConnContext) Value(key any) any {
	switch key {
	case constant.OpUserID:
		return c.GetUserID()
	case constant.OperationID:
		return c.GetOperationID()
	case constant.ConnID:
		return c.GetConnID()
	case constant.OpUserPlatform:
		return constant.PlatformIDToName(utils.StringToInt(c.GetPlatformID()))
	case constant.RemoteAddr:
		return c.RemoteAddr
	default:
		return ""
	}
}

func newContext(remoteAddr string, userID string) *UserConnContext {
	return &UserConnContext{
		//RespWriter: respWriter,
		//Req:        req,
		//Path:       req.URL.Path,
		//Method:     req.Method,
		userID:     userID,
		RemoteAddr: remoteAddr,
		ConnID:     utils.Md5(remoteAddr + "_" + strconv.Itoa(int(utils.GetCurrentTimestampByMill()))),
	}
}

func newTempContext() *UserConnContext {
	return &UserConnContext{
		//Req: &http.Request{URL: &url.URL{}},
	}
}

func (c *UserConnContext) GetRemoteAddr() string {
	return c.RemoteAddr
}

func (c *UserConnContext) GetConnID() string {
	return c.ConnID
}

func (c *UserConnContext) GetUserID() string {
	//return c.Req.URL.Query().Get(WsUserID)
	return c.userID
}

func (c *UserConnContext) GetPlatformID() string {
	//return c.Req.URL.Query().Get(PlatformID)
	return c.platformID
}

func (c *UserConnContext) GetOperationID() string {
	//return c.Req.URL.Query().Get(OperationID)
	return c.operationID
}

func (c *UserConnContext) SetOperationID(operationID string) {
	//values := c.Req.URL.Query()
	//values.Set(OperationID, operationID)
	//c.Req.URL.RawQuery = values.Encode()
	c.operationID = operationID
}

func (c *UserConnContext) GetToken() string {
	//return c.Req.URL.Query().Get(Token)
	return c.token
}

func (c *UserConnContext) SetToken(token string) {
	//c.Req.URL.RawQuery = Token + "=" + token
	c.token = token
}

func (c *UserConnContext) GetBackground() bool {
	//b, err := strconv.ParseBool(c.Req.URL.Query().Get(BackgroundStatus))
	//if err != nil {
	//	return false
	//} else {
	//	return b
	//}

	return c.isBackground
}

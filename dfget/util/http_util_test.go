/*
 * Copyright 1999-2018 Alibaba Group.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/go-check/check"
	"github.com/valyala/fasthttp"
)

type HTTPUtilTestSuite struct {
	host string
	ln   net.Listener
}

func init() {
	rand.Seed(time.Now().Unix())
	check.Suite(&HTTPUtilTestSuite{})
}

func (s *HTTPUtilTestSuite) SetUpSuite(c *check.C) {
	port := rand.Intn(1000) + 63000
	s.host = fmt.Sprintf("127.0.0.1:%d", port)

	s.ln, _ = net.Listen("tcp", s.host)
	go fasthttp.Serve(s.ln, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType(ApplicationJSONUtf8Value)
		ctx.SetStatusCode(fasthttp.StatusOK)
		req := &testJSONReq{}
		json.Unmarshal(ctx.Request.Body(), req)
		res := testJSONRes{
			Sum: req.A + req.B,
		}
		resByte, _ := json.Marshal(res)
		ctx.SetBody(resByte)
		time.Sleep(50 * time.Millisecond)
	})
}

func (s *HTTPUtilTestSuite) TearDownSuite(c *check.C) {
	s.ln.Close()
}

func (s *HTTPUtilTestSuite) TestPostJson(c *check.C) {
	var checkOk = func(code int, body []byte, e error, sum int) {
		c.Assert(e, check.IsNil)
		c.Assert(code, check.Equals, fasthttp.StatusOK)

		var res = &testJSONRes{}
		e = json.Unmarshal(body, res)
		c.Check(e, check.IsNil)
		c.Check(res.Sum, check.Equals, sum)
	}

	code, body, e := PostJSON("http://"+s.host, req(1, 2), 55*time.Millisecond)
	checkOk(code, body, e, 3)

	code, body, e = PostJSON("http://"+s.host, req(1, 2), 50*time.Millisecond)
	c.Assert(e, check.NotNil)
	c.Assert(e.Error(), check.Equals, "timeout")

	code, body, e = PostJSON("http://"+s.host, req(2, 3), 0)
	checkOk(code, body, e, 5)

	code, body, e = PostJSON("http://"+s.host, nil, 0)
	checkOk(code, body, e, 0)
}

func req(x int, y int) *testJSONReq {
	return &testJSONReq{x, y}
}

type testJSONReq struct {
	A int
	B int
}

type testJSONRes struct {
	Sum int
}

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/apm-server/internal/beater/request"
)

func TestTimeoutMiddleware(t *testing.T) {
	test := func(t *testing.T, handler request.Handler) {
		var err error
		h, err := TimeoutMiddleware()(handler)
		assert.NoError(t, err)

		c := request.NewContext()
		r, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)
		c.Reset(httptest.NewRecorder(), r)
		h(c)

		assert.Equal(t, http.StatusServiceUnavailable, c.Result.StatusCode)
	}
	t.Run("Cancelled", func(t *testing.T) {
		test(t, request.Handler(func(c *request.Context) {
			ctx := c.Request.Context()
			ctx, cancel := context.WithCancel(ctx)
			r := c.Request.WithContext(ctx)
			c.Request = r
			cancel()
		}))
	})
	t.Run("DeadlineExceeded", func(t *testing.T) {
		var cancel func()
		defer func() {
			if cancel != nil {
				cancel()
			}
		}()
		test(t, request.Handler(func(c *request.Context) {
			ctx := c.Request.Context()
			ctx, cancel = context.WithTimeout(ctx, time.Nanosecond)
			r := c.Request.WithContext(ctx)
			c.Request = r
			time.Sleep(time.Second)
		}))
	})
}

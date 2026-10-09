// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/pkg"
	"github.com/ory/x/configx"
)

func TestStartCourier(t *testing.T) {
	t.Run("case=without metrics", func(t *testing.T) {
		_, r := pkg.NewFastRegistryWithMocks(t)
		go func() { _ = StartCourier(t.Context(), r) }()
		require.Equal(t, r.Config().CourierExposeMetricsPort(t.Context()), 0)
	})

	t.Run("case=with metrics", func(t *testing.T) {
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		_, r := pkg.NewFastRegistryWithMocks(t, configx.WithValue("expose-metrics-port", port))
		go func() { _ = StartCourier(t.Context(), r) }()
		transport := http.DefaultTransport.(*http.Transport).Clone()
		t.Cleanup(transport.CloseIdleConnections)
		client := &http.Client{Transport: transport}
		url := fmt.Sprintf("http://127.0.0.1:%d/metrics/prometheus", port)
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			res, err := client.Get(url)
			if !assert.NoError(t, err) {
				return
			}
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.Equal(t, 200, res.StatusCode)
		}, 10*time.Second, 10*time.Millisecond)
	})
}

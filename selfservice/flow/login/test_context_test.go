// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
)

func TestFlow_IsTest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tf   sql.NullBool
		want bool
	}{
		{name: "null", tf: sql.NullBool{Valid: false}, want: false},
		{name: "false", tf: sql.NullBool{Valid: true, Bool: false}, want: false},
		{name: "true", tf: sql.NullBool{Valid: true, Bool: true}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := &login.Flow{TestFlow: tc.tf}
			assert.Equal(t, tc.want, f.IsTest())
		})
	}
}

func TestFlow_LoadTestContext_NonTestFlow(t *testing.T) {
	t.Parallel()
	f := &login.Flow{
		TestFlow:        sql.NullBool{Valid: false},
		InternalContext: sqlxx.JSONRawMessage(`{"test":{"provider_id":"google"},"foo":"bar"}`),
	}
	require.NoError(t, f.LoadTestContext())
	assert.Nil(t, f.TestContext)
}

func TestFlow_LoadTestContext_EmptyInternalContext(t *testing.T) {
	t.Parallel()
	f := &login.Flow{
		TestFlow:        sql.NullBool{Valid: true, Bool: true},
		InternalContext: nil,
	}
	require.NoError(t, f.LoadTestContext())
	assert.Nil(t, f.TestContext)
}

func TestFlow_LoadTestContext_RoundTrip(t *testing.T) {
	t.Parallel()
	tc := &login.TestContext{
		ProviderID: "google",
		DebugPayload: &login.DebugPayload{
			IDTokenClaims: map[string]any{"sub": "x"},
		},
	}
	f := &login.Flow{InternalContext: sqlxx.JSONRawMessage(`{}`)}
	require.NoError(t, f.SetTestContext(tc))
	assert.True(t, f.IsTest())

	// Simulate a round-trip through persistence by clearing the derived
	// field and reloading from InternalContext.
	f.TestContext = nil
	require.NoError(t, f.LoadTestContext())
	require.NotNil(t, f.TestContext)
	assert.Equal(t, "google", f.TestContext.ProviderID)
	require.NotNil(t, f.TestContext.DebugPayload)
	assert.Equal(t, "x", f.TestContext.DebugPayload.IDTokenClaims["sub"])
}

func TestFlow_SetTestContext_Nil_RemovesKey(t *testing.T) {
	t.Parallel()
	f := &login.Flow{
		InternalContext: sqlxx.JSONRawMessage(`{"test":{"provider_id":"google"},"foo":"bar"}`),
		TestFlow:        sql.NullBool{Valid: true, Bool: true},
	}
	require.NoError(t, f.SetTestContext(nil))

	assert.False(t, f.IsTest())
	require.NoError(t, f.LoadTestContext())
	assert.Nil(t, f.TestContext)
	assert.Equal(t, "bar", gjson.GetBytes(f.InternalContext, "foo").String())
	assert.False(t, gjson.GetBytes(f.InternalContext, "test").Exists())
}

func TestFlow_SetTestContext_Nil_EmptyInternalContext(t *testing.T) {
	t.Parallel()
	f := &login.Flow{InternalContext: nil}
	require.NoError(t, f.SetTestContext(nil))
	assert.Equal(t, sql.NullBool{Valid: true, Bool: false}, f.TestFlow)
	assert.True(t, len(f.InternalContext) == 0 || string(f.InternalContext) == "{}")
}

func TestFlow_SetTestContext_PreservesSiblingKeys(t *testing.T) {
	t.Parallel()
	f := &login.Flow{
		InternalContext: sqlxx.JSONRawMessage(`{"foo":"bar","baz":1}`),
	}
	require.NoError(t, f.SetTestContext(&login.TestContext{ProviderID: "x"}))

	assert.Equal(t, "bar", gjson.GetBytes(f.InternalContext, "foo").String())
	assert.Equal(t, int64(1), gjson.GetBytes(f.InternalContext, "baz").Int())
	assert.Equal(t, "x", gjson.GetBytes(f.InternalContext, "test.provider_id").String())
}

// newFlowForMarshal returns a minimal well-formed login flow suitable for
// MarshalJSON testing.
func newFlowForMarshal(t *testing.T) *login.Flow {
	t.Helper()
	return &login.Flow{
		ID:              x.NewUUID(),
		UI:              container.New(""),
		InternalContext: sqlxx.JSONRawMessage(`{}`),
	}
}

func TestFlow_MarshalJSON_TestFlowIncludesContext(t *testing.T) {
	t.Parallel()
	f := newFlowForMarshal(t)
	require.NoError(t, f.SetTestContext(&login.TestContext{
		ProviderID: "google",
		DebugPayload: &login.DebugPayload{
			IDTokenClaims: map[string]any{"sub": "alice"},
		},
	}))

	b, err := json.Marshal(*f)
	require.NoError(t, err)
	assert.Equal(t, "google", gjson.GetBytes(b, "test_context.provider_id").String())
	assert.Equal(t, "alice", gjson.GetBytes(b, "test_context.debug_payload.id_token_claims.sub").String())
}

func TestFlow_MarshalJSON_NonTestFlowOmitsContext(t *testing.T) {
	t.Parallel()
	f := newFlowForMarshal(t)

	b, err := json.Marshal(*f)
	require.NoError(t, err)
	assert.False(t, gjson.GetBytes(b, "test_context").Exists())
}

func TestFlow_MarshalJSON_UncapturedTestFlow(t *testing.T) {
	t.Parallel()
	f := newFlowForMarshal(t)
	require.NoError(t, f.SetTestContext(&login.TestContext{ProviderID: "google"}))

	b, err := json.Marshal(*f)
	require.NoError(t, err)
	assert.Equal(t, "google", gjson.GetBytes(b, "test_context.provider_id").String())
	assert.False(t, gjson.GetBytes(b, "test_context.debug_payload").Exists())
}

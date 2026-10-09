// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

// InternalContextTestKey is the JSON path inside a login flow's
// internal_context where test-mode data (provider id, debug payload) is
// stored. Using a dedicated sub-namespace keeps test fields out of the
// flow's first-class columns while still letting a single row carry them.
const InternalContextTestKey = "test"

// Admin-test extension of a login flow. Populated only for flows created by
// the admin test endpoint; included in the flow's API response so the admin
// UI can render the pre-scoped provider and (once captured) the debug
// round-trip result.
//
// swagger:model loginFlowTestContext
type TestContext struct {
	// The ID of the OIDC provider this test flow targets.
	//
	// required: true
	ProviderID string `json:"provider_id"`

	// Structured debug data captured during the provider round-trip.
	// Nil while the flow is still awaiting the callback.
	DebugPayload *DebugPayload `json:"debug_payload,omitempty"`
}

// Diagnostic payload captured during a test-mode OIDC round-trip.
// Contains the parsed claims, the Jsonnet mapper input and output, and any
// schema validation errors. Bearer tokens (id_token, access_token,
// refresh_token) are intentionally excluded to limit the blast radius of
// the debug payload leaking through audit logs or admin browsers.
//
// swagger:model loginFlowTestDebugPayload
type DebugPayload struct {
	// Claims extracted from the ID token.
	IDTokenClaims map[string]any `json:"id_token_claims,omitempty"`

	// Claims returned from the provider's userinfo endpoint, if any.
	Userinfo map[string]any `json:"userinfo,omitempty"`

	// URL of the Jsonnet mapper that was executed on the claims.
	JsonnetMapperURL string `json:"jsonnet_mapper_url,omitempty"`

	// Input JSON that was fed into the Jsonnet mapper.
	JsonnetInput map[string]any `json:"jsonnet_input,omitempty"`

	// Output JSON returned by the Jsonnet mapper.
	JsonnetOutput map[string]any `json:"jsonnet_output,omitempty"`

	// Anything the Jsonnet mapper wrote to standard error.
	JsonnetStderr string `json:"jsonnet_stderr,omitempty"`

	// Identity-schema validation errors produced from the mapped traits.
	SchemaValidationErrors []SchemaValidationError `json:"schema_validation_errors,omitempty"`

	// A classified error if any step of the round-trip failed.
	Error *StepError `json:"error,omitempty"`
}

// Classified failure of a step in an OIDC test round-trip.
// Populated when any step (token exchange, claims decode, Jsonnet
// evaluation, schema validation) cannot complete.
//
// swagger:model loginFlowTestStepError
type StepError struct {
	// Machine-readable identifier of the failed step (for example,
	// "token_exchange" or "schema_validate").
	Step string `json:"step"`

	// Short classification of the failure cause (for example,
	// "access_denied" or "traits_invalid").
	Reason string `json:"reason"`

	// Human-readable message describing the failure.
	Message string `json:"message"`
}

// One identity-schema validation failure recorded while evaluating the
// traits produced by the Jsonnet mapper.
//
// swagger:model loginFlowTestSchemaValidationError
type SchemaValidationError struct {
	// JSON pointer to the field that failed validation.
	Path string `json:"path"`

	// Human-readable description of the validation failure.
	Message string `json:"message"`
}

// IsTest reports whether this login flow is an admin-created test flow. The
// dedicated column lets persistence filter without parsing internal_context.
// A NULL or false test_flow column both read as false.
func (f *Flow) IsTest() bool { return f.TestFlow.Valid && f.TestFlow.Bool }

// LoadTestContext parses internal_context into the derived TestContext field
// so the flow's API response exposes it. A non-test flow or a flow with no
// test payload results in TestContext being nil.
func (f *Flow) LoadTestContext() error {
	if !f.IsTest() || len(f.InternalContext) == 0 {
		f.TestContext = nil
		return nil
	}
	var shell struct {
		Test *TestContext `json:"test"`
	}
	if err := json.Unmarshal(f.InternalContext, &shell); err != nil {
		return errors.WithStack(err)
	}
	f.TestContext = shell.Test
	return nil
}

// SetTestContext writes the derived TestContext back into internal_context
// and flips the TestFlow column. Call this after mutating f.TestContext (for
// example, when capturing a debug payload on the callback).
func (f *Flow) SetTestContext(tc *TestContext) error {
	f.TestContext = tc
	if tc == nil {
		f.TestFlow = sql.NullBool{Valid: true, Bool: false}
		if len(f.InternalContext) == 0 {
			return nil
		}
		updated, err := sjson.DeleteBytes(f.InternalContext, InternalContextTestKey)
		if err != nil {
			return errors.WithStack(err)
		}
		f.InternalContext = updated
		return nil
	}

	f.TestFlow = sql.NullBool{Valid: true, Bool: true}
	if len(f.InternalContext) == 0 {
		f.InternalContext = []byte("{}")
	}
	encoded, err := json.Marshal(tc)
	if err != nil {
		return errors.WithStack(err)
	}
	updated, err := sjson.SetRawBytes(f.InternalContext, InternalContextTestKey, encoded)
	if err != nil {
		return errors.WithStack(err)
	}
	f.InternalContext = updated
	return nil
}

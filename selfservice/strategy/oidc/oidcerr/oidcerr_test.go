// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidcerr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/herodot"
	"github.com/ory/kratos/selfservice/strategy/oidc/oidcerr"
)

func TestUnwrap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStep   oidcerr.Step
		wantReason string
	}{
		{
			name:       "state validate",
			err:        oidcerr.Wrap(oidcerr.StepStateValidate, herodot.ErrBadRequest().WithReason("invalid state")),
			wantStep:   oidcerr.StepStateValidate,
			wantReason: "state_invalid",
		},
		{
			name:       "provider denied",
			err:        oidcerr.Wrap(oidcerr.StepProviderDenied, herodot.ErrBadRequest().WithReason("access_denied: user refused")),
			wantStep:   oidcerr.StepProviderDenied,
			wantReason: "access_denied",
		},
		{
			name:       "token exchange",
			err:        oidcerr.Wrap(oidcerr.StepTokenExchange, errors.New("oauth2: cannot fetch token")),
			wantStep:   oidcerr.StepTokenExchange,
			wantReason: "exchange_failed",
		},
		{
			name:       "id_token verify",
			err:        oidcerr.Wrap(oidcerr.StepIDTokenVerify, herodot.ErrForbidden().WithReason("Could not verify id_token")),
			wantStep:   oidcerr.StepIDTokenVerify,
			wantReason: "verification_failed",
		},
		{
			name:       "claims decode",
			err:        oidcerr.Wrap(oidcerr.StepClaimsDecode, herodot.ErrUpstreamError().WithReason("provider did not return a subject")),
			wantStep:   oidcerr.StepClaimsDecode,
			wantReason: "claims_invalid",
		},
		{
			name:       "explicit callback bucket",
			err:        oidcerr.Wrap(oidcerr.StepCallback, errors.New("missing code")),
			wantStep:   oidcerr.StepCallback,
			wantReason: "callback_failed",
		},
		{
			name:       "untagged error falls back to callback",
			err:        errors.New("some unrelated failure"),
			wantStep:   oidcerr.StepCallback,
			wantReason: "callback_failed",
		},
		{
			name:       "untagged herodot error falls back to callback",
			err:        herodot.ErrInternalServerError().WithReason("oops"),
			wantStep:   oidcerr.StepCallback,
			wantReason: "callback_failed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := oidcerr.Unwrap(tc.err)
			assert.Equal(t, tc.wantStep, e.Step)
			assert.Equal(t, tc.wantReason, e.Step.Reason())
		})
	}
}

func TestWrap_NilPassthrough(t *testing.T) {
	t.Parallel()
	assert.Nil(t, oidcerr.Wrap(oidcerr.StepCallback, nil))
}

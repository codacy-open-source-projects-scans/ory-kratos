// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// Package oidcerr defines structured errors raised by the OIDC strategy.
//
// The Step type and its constants name the high-level callback phase that
// failed. The test login flow's debug renderer buckets failures by Step;
// non-test flows ignore it.
package oidcerr

import "errors"

// Step identifies the OIDC callback phase that produced an error.
type Step string

const (
	StepStateValidate  Step = "state_validate"
	StepProviderDenied Step = "provider_denied"
	StepTokenExchange  Step = "token_exchange"
	StepIDTokenVerify  Step = "id_token_verify"
	StepClaimsDecode   Step = "claims_decode"
	StepCallback       Step = "callback"
)

// stepReasons crosswalks a Step to the reason string surfaced in the test
// debug payload's StepError.Reason field.
var stepReasons = map[Step]string{
	StepStateValidate:  "state_invalid",
	StepProviderDenied: "access_denied",
	StepTokenExchange:  "exchange_failed",
	StepIDTokenVerify:  "verification_failed",
	StepClaimsDecode:   "claims_invalid",
	StepCallback:       "callback_failed",
}

// Reason returns the reason string surfaced in the test debug payload's
// StepError.Reason field. Defined on Step (not *CallbackStepError) because
// herodot.ReasonCarrier picks up any error type with a Reason() method and
// would clobber the herodot DefaultError reason on the wire.
func (s Step) Reason() string {
	return stepReasons[s]
}

// CallbackStepError tags an OIDC callback error with the step it belongs to.
// It wraps the underlying error transparently: errors.As walks past it to
// find the inner *herodot.DefaultError, so existing handlers and writers
// behave identically.
type CallbackStepError struct {
	Step Step
	err  error
}

func (e *CallbackStepError) Error() string { return e.err.Error() }
func (e *CallbackStepError) Unwrap() error { return e.err }

// Wrap tags err with a step hint. Pass-through for nil. The wrap is the
// outermost layer so callers that errors.As on *herodot.DefaultError keep
// working without changes.
func Wrap(step Step, err error) error {
	if err == nil {
		return nil
	}
	return &CallbackStepError{Step: step, err: err}
}

// Unwrap walks err's chain looking for a *CallbackStepError tag. It returns
// a pointer to the tagged error if found, or a fresh fallback tagged with
// StepCallback otherwise. The returned value is never nil — callers can
// always read e.Step / e.Step.Reason() without a nil check.
func Unwrap(err error) *CallbackStepError {
	if se, ok := errors.AsType[*CallbackStepError](err); ok {
		return se
	}
	return &CallbackStepError{Step: StepCallback, err: err}
}

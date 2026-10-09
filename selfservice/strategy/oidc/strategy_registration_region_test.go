// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/region"
)

func TestSetRegion(t *testing.T) {
	t.Parallel()

	t.Run("case=sets region from mapper output", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"region": "eu-central", "traits": {"subject": "alice@example.com"}}}`
		require.NoError(t, setRegion(evaluated, i))
		assert.Equal(t, region.EUCentral, i.Region)
	})

	t.Run("case=leaves region empty when mapper output omits region", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"traits": {"subject": "alice@example.com"}}}`
		require.NoError(t, setRegion(evaluated, i))
		assert.Equal(t, region.Region(""), i.Region)
	})

	t.Run("case=rejects empty region", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"region": "", "traits": {"subject": "alice@example.com"}}}`
		err := setRegion(evaluated, i)
		require.Error(t, err)
	})

	t.Run("case=rejects unknown region value", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"region": "not-a-region", "traits": {"subject": "alice@example.com"}}}`
		err := setRegion(evaluated, i)
		require.Error(t, err)
		assert.Equal(t, region.Region(""), i.Region)
	})

	t.Run("case=rejects non-string region", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"region": 42, "traits": {"subject": "alice@example.com"}}}`
		err := setRegion(evaluated, i)
		require.Error(t, err)
	})

	t.Run("case=accepts super-region", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{}
		evaluated := `{"identity": {"region": "eu", "traits": {"subject": "alice@example.com"}}}`
		require.NoError(t, setRegion(evaluated, i))
		assert.Equal(t, region.EU, i.Region)
	})
}

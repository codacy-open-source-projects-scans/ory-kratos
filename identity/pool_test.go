// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
)

func TestCreateIdentitiesOptions_WithExtraColumns(t *testing.T) {
	o := identity.NewCreateIdentitiesOptions([]identity.CreateIdentitiesModifier{
		identity.WithExtraColumns([]identity.ExtraColumn{{K: "crdb_region", V: "gcp-europe-west3"}}),
	})
	got := o.ExtraColumns
	require.Len(t, got, 1)
	assert.Equal(t, "crdb_region", got[0].K)
	assert.Equal(t, "gcp-europe-west3", got[0].V)
}

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/region"
)

func TestIdentity_RegionJSONRoundTrip(t *testing.T) {
	i := &identity.Identity{ID: uuid.Must(uuid.NewV4()), SchemaID: "default", Region: region.EUCentral}
	b, err := json.Marshal(i)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"region":"eu-central"`)

	var decoded identity.Identity
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, region.EUCentral, decoded.Region)
}

func TestIdentity_RegionJSONOmitEmpty(t *testing.T) {
	i := &identity.Identity{ID: uuid.Must(uuid.NewV4()), SchemaID: "default"}
	b, err := json.Marshal(i)
	require.NoError(t, err)
	assert.NotContains(t, string(b), `"region"`)
}

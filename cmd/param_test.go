package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryParamCmd(t *testing.T) {
	cmd := QueryParamCmd()

	assert.Equal(t, "param", cmd.Use)
	assert.Equal(t, "Query a governance parameter value", cmd.Short)

	cidFlag := cmd.Flags().Lookup("cid")
	assert.NotNil(t, cidFlag)
	assert.Equal(t, "c", cidFlag.Shorthand)
}

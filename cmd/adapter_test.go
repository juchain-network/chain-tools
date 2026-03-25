package cmd

import (
	"bytes"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackVoteProposalData(t *testing.T) {
	proposalID := "0x" + strings.Repeat("11", 32)

	data, err := packVoteProposalData(proposalID, true)
	require.NoError(t, err)
	require.Len(t, data, 68)

	expectedSelector := crypto.Keccak256([]byte("voteProposal(bytes32,bool)"))[:4]
	assert.True(t, bytes.Equal(expectedSelector, data[:4]))
}

func TestPackCreateOrEditValidatorData(t *testing.T) {
	feeAddr := "0x1000000000000000000000000000000000000001"
	signerAddr := "0x2000000000000000000000000000000000000002"

	t.Run("fee only overload", func(t *testing.T) {
		data, methodName, moniker, identity, website, email, details, err := packCreateOrEditValidatorData(
			feeAddr,
			"",
			"",
			"id",
			"https://example.com",
			"ops@example.com",
			"",
		)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		expectedSelector := crypto.Keccak256([]byte("createOrEditValidator(address,string,string,string,string,string)"))[:4]
		assert.True(t, bytes.Equal(expectedSelector, data[:4]))
		assert.Equal(t, "createOrEditValidator0", methodName)
		assert.Equal(t, "validator", moniker)
		assert.Equal(t, "id", identity)
		assert.Equal(t, "https://example.com", website)
		assert.Equal(t, "ops@example.com", email)
		assert.Equal(t, "Validator node", details)
	})

	t.Run("signer overload", func(t *testing.T) {
		data, methodName, moniker, identity, website, email, details, err := packCreateOrEditValidatorData(
			feeAddr,
			signerAddr,
			"my-validator",
			"",
			"",
			"",
			"rotating signer",
		)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		expectedSelector := crypto.Keccak256([]byte("createOrEditValidator(address,address,string,string,string,string,string)"))[:4]
		assert.True(t, bytes.Equal(expectedSelector, data[:4]))
		assert.Equal(t, "createOrEditValidator", methodName)
		assert.Equal(t, "my-validator", moniker)
		assert.Equal(t, "", identity)
		assert.Equal(t, "", website)
		assert.Equal(t, "", email)
		assert.Equal(t, "rotating signer", details)
	})
}

func TestIsFutureEffective(t *testing.T) {
	assert.True(t, isFutureEffective(100, big.NewInt(101), true))
	assert.False(t, isFutureEffective(100, big.NewInt(100), true))
	assert.False(t, isFutureEffective(100, big.NewInt(99), true))
	assert.False(t, isFutureEffective(100, nil, true))
	assert.False(t, isFutureEffective(100, big.NewInt(101), false))
}

func TestLoadDoubleSignHeader(t *testing.T) {
	t.Run("from hex flag", func(t *testing.T) {
		header, err := loadDoubleSignHeader("0x1234abcd", "", "header1")
		require.NoError(t, err)
		assert.Equal(t, []byte{0x12, 0x34, 0xab, 0xcd}, header)
	})

	t.Run("from hex file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "header.hex")
		require.NoError(t, os.WriteFile(path, []byte("0x0102"), 0o644))

		header, err := loadDoubleSignHeader("", path, "header1")
		require.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02}, header)
	})

	t.Run("from binary file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "header.bin")
		require.NoError(t, os.WriteFile(path, []byte{0x99, 0xaa, 0xbb}, 0o644))

		header, err := loadDoubleSignHeader("", path, "header2")
		require.NoError(t, err)
		assert.Equal(t, []byte{0x99, 0xaa, 0xbb}, header)
	})

	t.Run("reject both sources", func(t *testing.T) {
		_, err := loadDoubleSignHeader("0x12", "header.bin", "header1")
		require.Error(t, err)
	})

	t.Run("reject empty", func(t *testing.T) {
		_, err := loadDoubleSignHeader("", "", "header2")
		require.Error(t, err)
	})
}

func TestPackSubmitDoubleSignEvidenceData(t *testing.T) {
	data, err := packSubmitDoubleSignEvidenceData([]byte{0x01, 0x02}, []byte{0x03, 0x04})
	require.NoError(t, err)
	require.NotEmpty(t, data)

	expectedSelector := crypto.Keccak256([]byte("submitDoubleSignEvidence(bytes,bytes)"))[:4]
	assert.True(t, bytes.Equal(expectedSelector, data[:4]))
}

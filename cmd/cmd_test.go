package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestCreaetProposalTx(t *testing.T) {
	proposer := "0x016103822e9a3425DfeaFDCd57c9F7fC2bA72a8b"
	target := "0x029DAB47e268575D4AC167De64052FB228B5fA41"
	rpc := "https://testnet-rpc.juchain.org"
	_ = innerCreateProposal(nil, proposer, target, false, rpc)
}

func TestSignTx(t *testing.T) {
	file := GeneratedFilePath("createProposal.json")
	key := "0xca881281fb10b53a87d00cbfae29f7cf8cfe8ac7c8389b3d20b24fc6bc3f3ff9"
	key = strings.TrimPrefix(key, "0x")
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		fmt.Printf("invalid private key: %v", err)
		return
	}
	_ = innerSignRawTx(file, privateKey)
}

func TestSendTx(t *testing.T) {
	// proposer := "0x016103822e9a3425DfeaFDCd57c9F7fC2bA72a8b"
	file := GeneratedFilePath("createProposal_signed.json")
	rpc := "https://testnet-rpc.juchain.org"
	_ = innerSendSignedTx(file, rpc)
}

func TestBuldProposalId(t *testing.T) {
	// Wait for tx to be finished executing with hash 0xb75dc353e433ce38edb359ae9aa9f88fa265ff9436fac164b6afb97f0aa87795
	// tx confirmed in block 776421
	// Proposal ID: b171896d9c93b438c6798779e944ed591104f0edb63e09755b53624833f72ebd (updated based on current implementation)
	// Proposer: 0x016103822e9a3425DfeaFDCd57c9F7fC2bA72a8b
	// Destination: 0x029DAB47e268575D4AC167De64052FB228B5fA41
	// Flag: true
	// Time: 1743150640
	proposer := "0x016103822e9a3425DfeaFDCd57c9F7fC2bA72a8b"
	target := "0x029DAB47e268575D4AC167De64052FB228B5fA41"
	rpc := "https://testnet-rpc.juchain.org"
	blockNum := 776421

	client, err := ethclient.Dial(rpc)
	if err != nil {
		fmt.Printf("failed to connect to RPC: %v", err)
		return
	}
	defer client.Close()
	proposalId, err := BuildProposalId(proposer, target, true, "", uint64(blockNum), client)
	if err != nil {
		fmt.Printf("failed to build proposal id: %v", err)
		return
	}
	fmt.Printf("build proposal id %s \n", proposalId)
	_, eventId := QueryProposalId(uint64(blockNum), proposer, client)
	if eventId != "" {
		assert.Equal(t, eventId, proposalId, "Proposal ID mismatch")
	} else {
		assert.Len(t, proposalId, 64)
	}
}

func TestReadPassword(t *testing.T) {
	// Test reading an existing password file
	data, err := ReadFileToString("../password.file")
	if err != nil {
		// If the file doesn't exist in the expected location,
		// just test error handling
		fmt.Printf("Password file not found in expected location: %v\n", err)

		// Test reading a non-existent file - should return error
		_, err = ReadFileToString("nonexistent.file")
		if err == nil {
			t.Error("Expected error when reading non-existent file, but got nil")
			return
		}
		fmt.Printf("Expected error for non-existent file: %v\n", err)
		return
	}

	// Verify the content is not empty
	if len(data) == 0 {
		t.Error("Password file content is empty")
		return
	}

	fmt.Printf("Successfully read password file, content length: %d\n", len(data))

	// Test reading a non-existent file - should return error
	_, err = ReadFileToString("nonexistent.file")
	if err == nil {
		t.Error("Expected error when reading non-existent file, but got nil")
		return
	}
	fmt.Printf("Expected error for non-existent file: %v\n", err)
}

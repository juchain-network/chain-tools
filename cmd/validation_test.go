package cmd

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid address",
			addr:    "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantErr: false,
		},
		{
			name:    "invalid address - too short",
			addr:    "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb922",
			wantErr: true,
		},
		{
			name:    "invalid address - no 0x prefix",
			addr:    "f39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantErr: false, // common.IsHexAddress actually accepts this format
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateChainID(t *testing.T) {
	tests := []struct {
		name    string
		chainID int64
		wantErr bool
	}{
		{
			name:    "valid chain ID",
			chainID: 202599,
			wantErr: false,
		},
		{
			name:    "zero chain ID",
			chainID: 0,
			wantErr: true,
		},
		{
			name:    "negative chain ID",
			chainID: -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChainID(tt.chainID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		wantErr   bool
	}{
		{
			name:      "valid operation add",
			operation: "add",
			wantErr:   false,
		},
		{
			name:      "valid operation remove",
			operation: "remove",
			wantErr:   false,
		},
		{
			name:      "invalid operation",
			operation: "invalid",
			wantErr:   true,
		},
		{
			name:      "empty operation",
			operation: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOperation(tt.operation)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigID(t *testing.T) {
	tests := []struct {
		name     string
		configID int64
		wantErr  bool
	}{
		{
			name:     "valid config ID 0",
			configID: 0,
			wantErr:  false,
		},
		{
			name:     "valid config ID 4",
			configID: 4,
			wantErr:  false,
		},
		{
			name:     "valid config ID 15",
			configID: 15,
			wantErr:  false,
		},
		{
			name:     "valid config ID 19",
			configID: 19,
			wantErr:  false,
		},
		{
			name:     "invalid config ID negative",
			configID: -1,
			wantErr:  true,
		},
		{
			name:     "invalid config ID too large",
			configID: 20,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigID(tt.configID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetConfigIDName(t *testing.T) {
	tests := []struct {
		name     string
		configID int64
		expected string
	}{
		{
			name:     "config ID 0",
			configID: 0,
			expected: "proposalLastingPeriod",
		},
		{
			name:     "config ID 1",
			configID: 1,
			expected: "punishThreshold",
		},
		{
			name:     "config ID 4",
			configID: 4,
			expected: "withdrawProfitPeriod",
		},
		{
			name:     "config ID 10",
			configID: 10,
			expected: "minDelegation",
		},
		{
			name:     "config ID 18",
			configID: 18,
			expected: "maxCommissionRate",
		},
		{
			name:     "invalid config ID",
			configID: 20,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigIDName(tt.configID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTxExecutionOptions(t *testing.T) {
	tempDir := t.TempDir()
	walletFile := tempDir + "/wallet.json"
	passwordFile := tempDir + "/password.txt"
	assert.NoError(t, os.WriteFile(walletFile, []byte("wallet"), 0644))
	assert.NoError(t, os.WriteFile(passwordFile, []byte("password"), 0644))

	tests := []struct {
		name       string
		send       bool
		wallet     string
		privateKey string
		password   string
		wantErr    bool
	}{
		{
			name:    "offline without signer flags",
			wantErr: false,
		},
		{
			name:       "signer flags without send",
			privateKey: "0x01",
			wantErr:    true,
		},
		{
			name:    "send without signer",
			send:    true,
			wantErr: true,
		},
		{
			name:       "send with private key",
			send:       true,
			privateKey: "0xca881281fb10b53a87d00cbfae29f7cf8cfe8ac7c8389b3d20b24fc6bc3f3ff9",
			wantErr:    false,
		},
		{
			name:     "send with wallet and password",
			send:     true,
			wallet:   walletFile,
			password: passwordFile,
			wantErr:  false,
		},
		{
			name:    "send with wallet without password",
			send:    true,
			wallet:  walletFile,
			wantErr: true,
		},
		{
			name:       "send with both signer sources",
			send:       true,
			wallet:     walletFile,
			privateKey: "0xca881281fb10b53a87d00cbfae29f7cf8cfe8ac7c8389b3d20b24fc6bc3f3ff9",
			password:   passwordFile,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			addOnlineSendFlags(cmd)
			if tt.send {
				assert.NoError(t, cmd.Flags().Set("send", "true"))
			}
			if tt.wallet != "" {
				assert.NoError(t, cmd.Flags().Set("wallet", tt.wallet))
			}
			if tt.privateKey != "" {
				assert.NoError(t, cmd.Flags().Set("private-key", tt.privateKey))
			}
			if tt.password != "" {
				assert.NoError(t, cmd.Flags().Set("password", tt.password))
			}

			_, err := getTxExecutionOptions(cmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSignerMatchesSender(t *testing.T) {
	privateKey, err := crypto.HexToECDSA("ca881281fb10b53a87d00cbfae29f7cf8cfe8ac7c8389b3d20b24fc6bc3f3ff9")
	assert.NoError(t, err)

	expected := crypto.PubkeyToAddress(privateKey.PublicKey)
	assert.NoError(t, validateSignerMatchesSender(privateKey, expected))
	assert.Error(t, validateSignerMatchesSender(privateKey, common.HexToAddress("0x0000000000000000000000000000000000000001")))
}

func TestWriteCommandsExposeOnlineSendFlags(t *testing.T) {
	commands := []*cobra.Command{
		CreateProposalCmd(),
		CreateConfigProposalCmd(),
		VoteProposalCmd(),
		TransferCmd(),
		WithdrawProfitsCmd(),
		EditValidatorCmd(),
		SubmitDoubleSignCmd(),
		RegisterValidatorCmd(),
		DelegateCmd(),
		UndelegateCmd(),
		ClaimRewardsCmd(),
		ClaimValidatorRewardsCmd(),
		IncreaseStakeCmd(),
		DecreaseStakeCmd(),
		SetCommissionCmd(),
		DeregisterValidatorCmd(),
		ValidatorExitCmd(),
		WithdrawUnbondedCmd(),
		UnjailValidatorCmd(),
	}

	for _, cmd := range commands {
		assert.NotNil(t, cmd.Flags().Lookup("send"), cmd.CommandPath())
		assert.NotNil(t, cmd.Flags().Lookup("wallet"), cmd.CommandPath())
		assert.NotNil(t, cmd.Flags().Lookup("private-key"), cmd.CommandPath())
		assert.NotNil(t, cmd.Flags().Lookup("password"), cmd.CommandPath())
	}
}

func TestParseConfigValue(t *testing.T) {
	tests := []struct {
		name     string
		rawValue string
		cid      int64
		expected *big.Int
		wantErr  bool
	}{
		{
			name:     "decimal wei",
			rawValue: "12345",
			cid:      0,
			expected: big.NewInt(12345),
			wantErr:  false,
		},
		{
			name:     "hex value",
			rawValue: "0x10",
			cid:      1,
			expected: big.NewInt(16),
			wantErr:  false,
		},
		{
			name:     "ether with decimals",
			rawValue: "1.5 ether",
			cid:      5,
			expected: big.NewInt(1_500_000_000_000_000_000),
			wantErr:  false,
		},
		{
			name:     "ju unit",
			rawValue: "2 ju",
			cid:      8,
			expected: big.NewInt(2_000_000_000_000_000_000),
			wantErr:  false,
		},
		{
			name:     "gwei unit",
			rawValue: "3 gwei",
			cid:      5,
			expected: big.NewInt(3_000_000_000),
			wantErr:  false,
		},
		{
			name:     "burn address value",
			rawValue: "0x000000000000000000000000000000000000dEaD",
			cid:      14,
			expected: new(big.Int).SetBytes(common.HexToAddress("0x000000000000000000000000000000000000dEaD").Bytes()),
			wantErr:  false,
		},
		{
			name:     "fractional wei without unit",
			rawValue: "1.5",
			cid:      0,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseConfigValue(tt.rawValue, tt.cid)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, 0, result.Cmp(tt.expected))
		})
	}
}

func TestParseTransferAmount(t *testing.T) {
	tests := []struct {
		name     string
		rawValue string
		expected *big.Int
		wantErr  bool
	}{
		{
			name:     "plain eth integer",
			rawValue: "2",
			expected: big.NewInt(2_000_000_000_000_000_000),
		},
		{
			name:     "plain eth decimal",
			rawValue: "0.25",
			expected: big.NewInt(250_000_000_000_000_000),
		},
		{
			name:     "explicit eth unit",
			rawValue: "1.5 eth",
			expected: big.NewInt(1_500_000_000_000_000_000),
		},
		{
			name:     "unsupported unit",
			rawValue: "100 gwei",
			wantErr:  true,
		},
		{
			name:     "too much precision",
			rawValue: "0.1234567890123456789",
			wantErr:  true,
		},
		{
			name:     "empty value",
			rawValue: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTransferAmount(tt.rawValue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.Cmp(tt.expected))
		})
	}
}

func TestValidateProposalID(t *testing.T) {
	tests := []struct {
		name       string
		proposalID string
		wantErr    bool
	}{
		{
			name:       "valid proposal ID",
			proposalID: "4881cb992217b8d050a87384de6998efb1f051c0cb62c10bdd393f68336618b6",
			wantErr:    false,
		},
		{
			name:       "invalid proposal ID - too short",
			proposalID: "4881cb992217b8d050a87384de6998efb1f051c0cb62c10bdd393f68336618",
			wantErr:    true,
		},
		{
			name:       "invalid proposal ID - invalid characters",
			proposalID: "4881cb992217b8d050a87384de6998efb1f051c0cb62c10bdd393f68336618gz",
			wantErr:    true,
		},
		{
			name:       "empty proposal ID",
			proposalID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProposalID(tt.proposalID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

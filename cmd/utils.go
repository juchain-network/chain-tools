package cmd

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"juchain.org/chain/tools/contracts"
)

// ValidateAddress validates Ethereum address format
func ValidateAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if !common.IsHexAddress(addr) {
		return fmt.Errorf("invalid address format: %s", addr)
	}
	return nil
}

// ValidateAddresses validates multiple addresses
func ValidateAddresses(addrs ...string) error {
	for _, addr := range addrs {
		if err := ValidateAddress(addr); err != nil {
			return err
		}
	}
	return nil
}

// ValidateChainID validates chain ID
func ValidateChainID(chainID int64) error {
	if chainID <= 0 {
		return fmt.Errorf("invalid chain ID: %d, must be positive", chainID)
	}
	return nil
}

// ValidateRPCURL validates RPC URL format
func ValidateRPCURL(url string) error {
	if url == "" {
		return fmt.Errorf("RPC URL cannot be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "ws://") && !strings.HasPrefix(url, "wss://") {
		return fmt.Errorf("invalid RPC URL format: %s", url)
	}
	return nil
}

// ValidateFile validates file existence
func ValidateFile(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filepath)
	}
	return nil
}

// GeneratedFilePath returns the default output path under the repository data directory.
func GeneratedFilePath(filename string) string {
	if filename == "" {
		return ""
	}
	if filepath.IsAbs(filename) || filepath.Dir(filename) != "." {
		return filename
	}
	return filepath.Join(DefaultDataDir, filename)
}

// EnsureParentDir creates the parent directory for a file if needed.
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// ResolveExistingFile prefers the provided path, but falls back to data/<name> for bare filenames.
func ResolveExistingFile(path string) string {
	if path == "" {
		return ""
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if filepath.IsAbs(path) || filepath.Dir(path) != "." {
		return path
	}
	candidate := GeneratedFilePath(path)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return path
}

// ValidateOperation validates operation type
func ValidateOperation(operation string) error {
	if operation != "add" && operation != "remove" {
		return fmt.Errorf("invalid operation: %s, must be 'add' or 'remove'", operation)
	}
	return nil
}

// ValidateConfigID validates configuration ID
func ValidateConfigID(cid int64) error {
	if cid < 0 || cid > 19 {
		return fmt.Errorf("invalid config ID: %d, must be 0-19", cid)
	}
	return nil
}

var (
	numberUnitRegexp = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*([a-zA-Z]+)?$`)
	hexValueRegexp   = regexp.MustCompile(`^[0-9a-f]+$`)
)

// ParseConfigValue parses config value with support for decimal, hex/address, and unit suffixes.
// Supported units: wei, gwei, ether, eth, ju (case-insensitive).
func ParseConfigValue(raw string, cid int64) (*big.Int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("config value cannot be empty")
	}

	lower := strings.ToLower(value)

	if strings.HasPrefix(lower, "0x") {
		return parseHexOrAddressValue(lower, cid)
	}

	// Address without 0x prefix is only valid for burnAddress (cid=14).
	if cid == 14 && common.IsHexAddress(value) {
		addr := common.HexToAddress(value)
		return addressToValue(addr)
	}

	matches := numberUnitRegexp.FindStringSubmatch(lower)
	if matches == nil {
		return nil, fmt.Errorf("invalid config value format: %s", raw)
	}

	number := matches[1]
	unit := matches[2]
	if unit == "" {
		if strings.Contains(number, ".") {
			return nil, fmt.Errorf("fractional wei requires a unit suffix (e.g. 1.5 ether)")
		}
		intValue, ok := new(big.Int).SetString(number, 10)
		if !ok {
			return nil, fmt.Errorf("invalid decimal value: %s", raw)
		}
		return intValue, nil
	}

	multiplier, ok := unitMultiplier(unit)
	if !ok {
		return nil, fmt.Errorf("unsupported unit: %s", unit)
	}

	rat, ok := new(big.Rat).SetString(number)
	if !ok {
		return nil, fmt.Errorf("invalid decimal value: %s", raw)
	}
	rat.Mul(rat, new(big.Rat).SetInt(multiplier))
	if !rat.IsInt() {
		return nil, fmt.Errorf("value has too much precision for unit %s", unit)
	}

	intValue := new(big.Int).Quo(rat.Num(), rat.Denom())
	return intValue, nil
}

// ParseTransferAmount parses a transfer amount expressed in ETH units.
// Bare numeric values such as "0.5" are treated as ETH and converted to wei.
func ParseTransferAmount(raw string) (*big.Int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("transfer amount cannot be empty")
	}

	matches := numberUnitRegexp.FindStringSubmatch(strings.ToLower(value))
	if matches == nil {
		return nil, fmt.Errorf("invalid transfer amount format: %s", raw)
	}

	unit := matches[2]
	switch unit {
	case "":
		return ParseConfigValue(value+" eth", 0)
	case "eth", "ether", "ju":
		return ParseConfigValue(value, 0)
	default:
		return nil, fmt.Errorf("transfer amount must use ETH units")
	}
}

func parseHexOrAddressValue(raw string, cid int64) (*big.Int, error) {
	hexStr := strings.TrimPrefix(raw, "0x")
	if hexStr == "" {
		return nil, fmt.Errorf("invalid hex value: %s", raw)
	}
	if !hexValueRegexp.MatchString(hexStr) {
		return nil, fmt.Errorf("invalid hex value: %s", raw)
	}

	if cid == 14 && len(hexStr) <= 40 {
		addr := common.HexToAddress(raw)
		return addressToValue(addr)
	}

	intValue, ok := new(big.Int).SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex value: %s", raw)
	}
	if cid == 14 && intValue.BitLen() > 160 {
		return nil, fmt.Errorf("burnAddress exceeds uint160 range")
	}
	return intValue, nil
}

func addressToValue(addr common.Address) (*big.Int, error) {
	value := new(big.Int).SetBytes(addr.Bytes())
	if value.Sign() <= 0 {
		return nil, fmt.Errorf("burnAddress must be non-zero")
	}
	return value, nil
}

func unitMultiplier(unit string) (*big.Int, bool) {
	switch strings.ToLower(unit) {
	case "wei":
		return big.NewInt(1), true
	case "gwei":
		return big.NewInt(1_000_000_000), true
	case "ether", "eth", "ju":
		return big.NewInt(1_000_000_000_000_000_000), true
	default:
		return nil, false
	}
}

// ValidateProposalID validates proposal ID format
func ValidateProposalID(proposalID string) error {
	if proposalID == "" {
		return fmt.Errorf("proposal ID cannot be empty")
	}

	// Allow 0x prefix
	cleanID := strings.TrimPrefix(proposalID, "0x")

	// Validate if it's a valid hexadecimal string
	if !regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(cleanID) {
		return fmt.Errorf("invalid proposal ID format: %s", proposalID)
	}
	if len(cleanID) != 64 {
		return fmt.Errorf("proposal ID must be 64 characters long: %s", proposalID)
	}
	return nil
}

// GetConfigIDName gets the name of the configuration ID
func GetConfigIDName(cid int64) string {
	switch cid {
	case 0:
		return "proposalLastingPeriod"
	case 1:
		return "punishThreshold"
	case 2:
		return "removeThreshold"
	case 3:
		return "decreaseRate"
	case 4:
		return "withdrawProfitPeriod"
	case 5:
		return "blockReward"
	case 6:
		return "unbondingPeriod"
	case 7:
		return "validatorUnjailPeriod"
	case 8:
		return "minValidatorStake"
	case 9:
		return "maxValidators"
	case 10:
		return "minDelegation"
	case 11:
		return "minUndelegation"
	case 12:
		return "doubleSignSlashAmount"
	case 13:
		return "doubleSignRewardAmount"
	case 14:
		return "burnAddress"
	case 15:
		return "doubleSignWindow"
	case 16:
		return "commissionUpdateCooldown"
	case 17:
		return "baseRewardRatio"
	case 18:
		return "maxCommissionRate"
	case 19:
		return "proposalCooldown"
	default:
		return "unknown"
	}
}

// PrintValidationError prints validation error message
func PrintValidationError(err error) {
	fmt.Printf("❌ Validation Error: %v\n", err)
}

// PrintSuccess prints success message
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintInfo prints information message
func PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintWarning prints warning message
func PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintError prints error message
func PrintError(message string, err error) {
	fmt.Printf("❌ %s: %v\n", message, err)
}

func EtherToWei(ether string) *big.Int {
	parts := strings.Split(ether, ".")
	if len(parts) == 1 {
		result := new(big.Int)
		result.SetString(parts[0], 10)
		return result.Mul(result, big.NewInt(1e18))
	}

	integerPart := new(big.Int)
	integerPart.SetString(parts[0], 10)
	integerPart.Mul(integerPart, big.NewInt(1e18))

	decimalPart := new(big.Int)
	decimalPart.SetString(parts[1], 10)

	zeros := 18 - len(parts[1])
	if zeros > 0 {
		decimalPart.Mul(decimalPart, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(zeros)), nil))
	}

	return integerPart.Add(integerPart, decimalPart)
}

func WeiToEther(wei *big.Int) string {
	eth := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return eth.Text('f', 18)
}

func GetCurrentBlockNumber(rpc string) (uint64, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		PrintError("Failed to connect to RPC endpoint", err)
		return 0, err
	}
	defer client.Close()

	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		PrintError("Failed to query current block number", err)
		return 0, err
	}
	return blockNumber, nil
}

func GetContractInstance(rpc string) (*contracts.Validators, *contracts.Staking, *contracts.Proposal, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		PrintError("Failed to connect to RPC endpoint", err)
		return nil, nil, nil, err
	}

	// Get validator information
	validatorInstance, err := contracts.NewValidators(common.HexToAddress(ValidatorContractAddr), client)
	if err != nil {
		PrintError("Failed to instantiate validator contract", err)
		return nil, nil, nil, err
	}
	// Connect to staking contract
	stakingInstance, err := contracts.NewStaking(common.HexToAddress(StakingContractAddr), client)
	if err != nil {
		PrintError("Failed to instantiate staking contract", err)
		return nil, nil, nil, err
	}
	// Connect to proposal contract
	proposalInstance, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		PrintError("Failed to instantiate proposal contract", err)
		return nil, nil, nil, err
	}
	return validatorInstance, stakingInstance, proposalInstance, nil
}

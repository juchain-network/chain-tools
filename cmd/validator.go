package cmd

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"juchain.org/chain/tools/contracts"
)

func ValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all validators",
		Run:   listValidators,
	}
	return cmd
}

func listValidators(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function

	// Validate RPC URL
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, stakingInstance, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	PrintInfo("Fetching validator information...")
	vals, err := validatorInstance.GetHighestValidators(&bind.CallOpts{})
	if err != nil {
		PrintError("Failed to get validators", err)
		return
	}

	if len(vals) == 0 {
		PrintInfo("No validators found")
		return
	}

	PrintInfo(fmt.Sprintf("Found %d validators:", len(vals)))
	for i, val := range vals {
		fmt.Printf("\n--- Validator %d ---\n", i+1)
		queryOneInfo(val.Hex(), validatorInstance, stakingInstance)
	}
}

func ValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query a validator by address",
		Run:   queryValidator,
	}
	queryValidatorFlags(cmd)
	return cmd
}

func queryValidatorFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "validator address to query")
	_ = cmd.MarkFlagRequired("addr")
}

func queryValidator(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	addr, _ := cmd.Flags().GetString("addr")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(addr); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, stakingInstance, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	PrintInfo(fmt.Sprintf("Querying validator information for: %s", addr))
	queryOneInfo(addr, validatorInstance, stakingInstance)
}

func queryOneInfo(addr string, validatorInstance *contracts.Validators, stakingInstance *contracts.Staking) {
	// Get basic validator info
	feeAddr, status, aacIncoming, totalJailedHB, lastWithdrawProfitsBlock, err := validatorInstance.GetValidatorInfo(&bind.CallOpts{}, common.HexToAddress(addr))
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get validator info for %s", addr), err)
		return
	}

	// Get validator description
	moniker, identity, website, email, details, err := validatorInstance.GetValidatorDescription(&bind.CallOpts{}, common.HexToAddress(addr))
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get validator description for %s", addr), err)
		return
	}

	// Get validator staking information
	validatorInfo, err := stakingInstance.GetValidatorInfo(&bind.CallOpts{}, common.HexToAddress(addr))
	if err != nil {
		PrintError("Failed to get validator staking info", err)
		return
	}

	fmt.Printf("Address: %s\n", addr)
	fmt.Printf("Fee Address: %s\n", feeAddr.Hex())
	// Print friendly status label instead of raw number
	fmt.Printf("Status: %s\n", formatValidatorStatus(uint64(status)))
	fmt.Printf("Self Stake: %s\n", WeiToEther(validatorInfo.SelfStake))
	fmt.Printf("Total Delegated: %s\n", WeiToEther(validatorInfo.TotalDelegated))
	fmt.Printf("Commission Rate: %.2f%%\n", float64(validatorInfo.CommissionRate.Int64())/100) // Commission rate is in basis points (0-10000)
	fmt.Printf("Block Rewards: %s\n", WeiToEther(aacIncoming))
	// totalJailedHB represents total penalized (forfeited) rewards
	fmt.Printf("Penalized Rewards: %s\n", WeiToEther(totalJailedHB))
	fmt.Printf("Last Withdraw Block: %s\n", lastWithdrawProfitsBlock.String())
	fmt.Printf("Moniker: %s\n", moniker)
	fmt.Printf("Identity: %s\n", identity)
	fmt.Printf("Website: %s\n", website)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Details: %s\n", details)
	if validatorInfo.AccumulatedRewards.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("Staking Rewards: %s\n", WeiToEther(validatorInfo.AccumulatedRewards))
	}
	if validatorInfo.IsJailed {
		fmt.Printf("Jailed: Yes, until block %s\n", validatorInfo.JailUntilBlock.String())
	}
	if validatorInfo.TotalClaimedRewards.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("Total Claimed Rewards: %s\n", WeiToEther(validatorInfo.TotalClaimedRewards))
		fmt.Printf("Last Claim Block: %s\n", validatorInfo.LastClaimBlock.String())
	}
}

// formatValidatorStatus converts numeric status to a concise, friendly label
// 0 -> "Not Exist ❌"; 1 -> "Active ✅"; 2 -> "Jailed ⚠️"; others -> "Unknown"
func formatValidatorStatus(status uint64) string {
	switch status {
	case 0:
		return "Not Exist ❌"
	case 1:
		return "Active ✅"
	case 2:
		return "Jailed ⚠️"
	default:
		return "Unknown"
	}
}

func WithdrawProfitsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim validator's reward",
		Run:   validatorClaim,
	}
	validatorClaimFlags(cmd)
	addOnlineSendFlags(cmd)
	return cmd
}

func validatorClaimFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "Caller address (transaction sender, required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address to claim rewards from (required)")
	_ = cmd.MarkFlagRequired("caller")
	_ = cmd.MarkFlagRequired("validator")
}

func validatorClaim(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	caller, _ := cmd.Flags().GetString("caller")
	validator, _ := cmd.Flags().GetString("validator")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(caller, validator); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo(fmt.Sprintf("Creating claim transaction: Caller=%s, Validator=%s", caller, validator))
	if err := innerValidatorClaim(cmd, caller, validator, rpc); err != nil {
		PrintError("Failed to create claim transaction", err)
		return
	}
}

func innerValidatorClaim(cmd *cobra.Command, caller string, validator string, rpc string) error {
	validatorAbi, err := abi.JSON(strings.NewReader(contracts.ValidatorsABI))
	if err != nil {
		return fmt.Errorf("failed to parse validator ABI: %w", err)
	}

	abiData, err := validatorAbi.Pack("withdrawProfits", common.HexToAddress(validator))
	if err != nil {
		return fmt.Errorf("failed to pack withdrawProfits data: %w", err)
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(caller),
		common.HexToAddress(ValidatorContractAddr),
		nil,
		abiData,
		rpc,
		WithdrawProfitsFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute withdraw transaction: %w", err)
	}

	printTxExecutionResult(result, "Withdraw profits transaction created successfully!")
	PrintWarning("Note: Withdrawal has minimum waiting period restrictions")
	return nil
}

// EditValidatorCmd creates command for editing validator information
func EditValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Create edit validator transaction",
		Long:  "Create a transaction to edit validator information including fee address and description",
		Run:   createEditValidatorTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")
	cmd.Flags().StringP("fee-addr", "f", "", "Fee address for receiving rewards (required)")
	cmd.Flags().StringP("moniker", "m", "", "Validator display name")
	cmd.Flags().StringP("identity", "i", "", "Validator identity (keybase signature)")
	cmd.Flags().StringP("website", "w", "", "Validator website URL")
	cmd.Flags().StringP("email", "e", "", "Validator contact email")
	cmd.Flags().StringP("details", "d", "", "Validator description details")

	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("fee-addr")
	addOnlineSendFlags(cmd)

	return cmd
}

func createEditValidatorTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")
	feeAddr, _ := cmd.Flags().GetString("fee-addr")
	moniker, _ := cmd.Flags().GetString("moniker")
	identity, _ := cmd.Flags().GetString("identity")
	website, _ := cmd.Flags().GetString("website")
	email, _ := cmd.Flags().GetString("email")
	details, _ := cmd.Flags().GetString("details")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(validatorAddr, feeAddr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating edit validator transaction")

	if err := innerCreateEditValidatorTx(cmd, validatorAddr, feeAddr, moniker, identity, website, email, details, rpc); err != nil {
		PrintError("Failed to create edit validator transaction", err)
		return
	}
}

func innerCreateEditValidatorTx(cmd *cobra.Command, validatorAddr, feeAddr, moniker, identity, website, email, details, rpc string) error {
	// Parse Validators contract ABI
	validatorsAbi, err := abi.JSON(strings.NewReader(contracts.ValidatorsABI))
	if err != nil {
		return fmt.Errorf("failed to parse validators ABI: %w", err)
	}

	// Use default values if not provided
	if moniker == "" {
		moniker = "validator"
	}
	if identity == "" {
		identity = ""
	}
	if website == "" {
		website = ""
	}
	if email == "" {
		email = ""
	}
	if details == "" {
		details = "Validator node"
	}

	abiData, err := validatorsAbi.Pack("createOrEditValidator",
		common.HexToAddress(feeAddr), moniker, identity, website, email, details)
	if err != nil {
		return fmt.Errorf("failed to pack createOrEditValidator data: %w", err)
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(validatorAddr),
		common.HexToAddress(ValidatorContractAddr),
		big.NewInt(0),
		abiData,
		rpc,
		EditValidatorFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute edit validator transaction: %w", err)
	}

	printTxExecutionResult(result, "Edit validator transaction created successfully!")
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	PrintInfo(fmt.Sprintf("Fee address: %s", feeAddr))
	PrintInfo(fmt.Sprintf("Moniker: %s", moniker))
	if identity != "" {
		PrintInfo(fmt.Sprintf("Identity: %s", identity))
	}
	if website != "" {
		PrintInfo(fmt.Sprintf("Website: %s", website))
	}
	if email != "" {
		PrintInfo(fmt.Sprintf("Email: %s", email))
	}
	if details != "" {
		PrintInfo(fmt.Sprintf("Details: %s", details))
	}
	return nil
}

package cmd

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"juchain.org/chain/tools/contracts"
)

// RegisterValidatorCmd creates command for validator registration
func RegisterValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-register",
		Short: "Create validator registration transaction",
		Long:  "Create a transaction to register as a validator by staking JU tokens and setting commission rate",
		Run:   createRegisterValidatorTx,
	}

	cmd.Flags().StringP("proposer", "p", "", "Proposer address (required)")
	cmd.Flags().StringP("stake-amount", "s", "", "Amount of JU to stake (required, minimum 10000)")
	cmd.Flags().StringP("commission-rate", "c", "", "Commission rate in basis points (0-10000, e.g., 500 = 5%)")

	_ = cmd.MarkFlagRequired("proposer")
	_ = cmd.MarkFlagRequired("stake-amount")
	_ = cmd.MarkFlagRequired("commission-rate")

	return cmd
}

// DelegateCmd creates command for delegation
func DelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "Create delegation transaction",
		Long:  "Create a transaction to delegate JU tokens to a validator",
		Run:   createDelegateTx,
	}

	cmd.Flags().StringP("delegator", "d", "", "Delegator address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address to delegate to (required)")
	cmd.Flags().StringP("amount", "s", "", "Amount of JU to delegate (required, minimum 1)")

	_ = cmd.MarkFlagRequired("delegator")
	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

// UndelegateCmd creates command for undelegation
func UndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate",
		Short: "Create undelegation transaction",
		Long:  "Create a transaction to start the 7-day unbonding process for delegated tokens",
		Run:   createUndelegateTx,
	}

	cmd.Flags().StringP("delegator", "d", "", "Delegator address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address to undelegate from (required)")
	cmd.Flags().StringP("amount", "s", "", "Amount of JU to undelegate (required)")

	_ = cmd.MarkFlagRequired("delegator")
	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

// ClaimRewardsCmd creates command for claiming rewards (for delegators)
func ClaimRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards",
		Short: "Create claim rewards transaction (for delegators)",
		Long:  "Create a transaction to claim accumulated staking rewards as a delegator",
		Run:   createClaimRewardsTx,
	}

	cmd.Flags().StringP("claimer", "c", "", "Claimer address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address to claim rewards from (required)")

	_ = cmd.MarkFlagRequired("claimer")
	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// ClaimValidatorRewardsCmd creates command for validators to claim their own rewards
func ClaimValidatorRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-validator-rewards",
		Short: "Create claim validator rewards transaction (for validators)",
		Long:  "Create a transaction for validators to claim their accumulated rewards",
		Run:   createClaimValidatorRewardsTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// QueryDelegationCmd creates command for querying delegation info
func QueryDelegationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-delegation",
		Short: "Query delegation information",
		Long:  "Get delegation details including amount and pending rewards",
		Run:   queryDelegationInfo,
	}

	cmd.Flags().StringP("delegator", "d", "", "Delegator address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("delegator")
	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// IncreaseStakeCmd creates command for increasing validator stake
func IncreaseStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake-increase",
		Short: "Increase validator stake amount",
		Long:  "Create a transaction to increase validator staking amount",
		Run:   createIncreaseStakeTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")
	cmd.Flags().StringP("amount", "s", "", "Amount of JU to add to stake (required)")

	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

// DecreaseStakeCmd creates command for decreasing validator stake
func DecreaseStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake-decrease",
		Short: "Decrease validator stake amount",
		Long:  "Create a transaction to decrease validator staking amount",
		Run:   createDecreaseStakeTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")
	cmd.Flags().StringP("amount", "s", "", "Amount of JU to remove from stake (required)")

	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

// SetCommissionCmd creates command for setting validator commission rate
func SetCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-commission",
		Short: "Set validator commission rate",
		Long:  "Create a transaction to set validator commission rate in basis points (0-10000)",
		Run:   createSetCommissionTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")
	cmd.Flags().StringP("rate", "r", "", "New commission rate in basis points (0-10000, required)")

	_ = cmd.MarkFlagRequired("validator")
	_ = cmd.MarkFlagRequired("rate")

	return cmd
}

// DeregisterValidatorCmd creates command for deregistering validator
func DeregisterValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-deregister",
		Short: "Deregister validator",
		Long:  "Create a transaction to deregister as a validator",
		Run:   createDeregisterValidatorTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// ValidatorExitCmd creates command for validator complete exit
func ValidatorExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-exit",
		Short: "Validator complete exit",
		Long:  "Create a transaction for validator to completely exit the network",
		Run:   createValidatorExitTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// WithdrawUnbondedCmd creates command for withdrawing unbonded stakes
func WithdrawUnbondedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-unbonded",
		Short: "Withdraw unbonded stakes",
		Long:  "Create a transaction to withdraw unbonded stakes",
		Run:   createWithdrawUnbondedTx,
	}

	cmd.Flags().StringP("claimer", "c", "", "Claimer address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address to withdraw from (required)")

	_ = cmd.MarkFlagRequired("claimer")
	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// UnjailValidatorCmd creates command for unjailing a validator
func UnjailValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-unjail",
		Short: "Unjail a validator",
		Long:  "Create a transaction to unjail a validator that has been jailed",
		Run:   createUnjailValidatorTx,
	}

	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// QueryAvailableUnbondedCmd creates command for querying available unbonded stakes
func QueryAvailableUnbondedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-unbonded",
		Short: "Query available unbonded stakes",
		Long:  "Get total available unbonded amount from unbonding delegations",
		Run:   queryAvailableUnbonded,
	}

	cmd.Flags().StringP("delegator", "d", "", "Delegator address (required)")
	cmd.Flags().StringP("validator", "v", "", "Validator address (required)")

	_ = cmd.MarkFlagRequired("delegator")
	_ = cmd.MarkFlagRequired("validator")

	return cmd
}

// Implementation functions

func createRegisterValidatorTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	proposer, _ := cmd.Flags().GetString("proposer")
	stakeAmountStr, _ := cmd.Flags().GetString("stake-amount")
	commissionRateStr, _ := cmd.Flags().GetString("commission-rate")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(proposer); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse amounts
	stakeAmount, ok := new(big.Int).SetString(stakeAmountStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid stake amount: %s", stakeAmountStr))
		return
	}

	// Convert to wei (multiply by 10^18)
	weiAmount := new(big.Int).Mul(stakeAmount, big.NewInt(1e18))

	commissionRate, ok := new(big.Int).SetString(commissionRateStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid commission rate: %s", commissionRateStr))
		return
	}

	// Validate commission rate (0-10000)
	if commissionRate.Cmp(big.NewInt(10000)) > 0 {
		PrintValidationError(fmt.Errorf("commission rate must be between 0 and 10000 (100%%)"))
		return
	}

	PrintInfo("Creating validator registration transaction")

	if err := innerCreateRegisterValidatorTx(proposer, weiAmount, commissionRate, rpc); err != nil {
		PrintError("Failed to create register validator transaction", err)
		return
	}
}

func innerCreateRegisterValidatorTx(proposer string, stakeAmount, commissionRate *big.Int, rpc string) error {
	outputFile := GeneratedFilePath(RegisterValidatorFile)

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("registerValidator", commissionRate)
	if err != nil {
		return fmt.Errorf("failed to pack registerValidator data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(proposer), common.HexToAddress(StakingContractAddr), stakeAmount, abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	PrintSuccess("Validator registration transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Stake amount: %s wei", stakeAmount.String()))
	PrintInfo(fmt.Sprintf("Commission rate: %s basis points", commissionRate.String()))
	return nil
}

func createDelegateTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	delegator, _ := cmd.Flags().GetString("delegator")
	validatorStr, _ := cmd.Flags().GetString("validator")
	amountStr, _ := cmd.Flags().GetString("amount")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(delegator, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse amount
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid amount: %s", amountStr))
		return
	}

	// Convert to wei
	weiAmount := new(big.Int).Mul(amount, big.NewInt(1e18))

	PrintInfo("Creating delegation transaction")

	if err := innerCreateDelegateTx(delegator, validatorStr, weiAmount, rpc); err != nil {
		PrintError("Failed to create delegate transaction", err)
		return
	}
}

func innerCreateDelegateTx(delegator, validator string, amount *big.Int, rpc string) error {
	outputFile := GeneratedFilePath(DelegateFile)

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("delegate", common.HexToAddress(validator))
	if err != nil {
		return fmt.Errorf("failed to pack delegate data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(delegator), common.HexToAddress(StakingContractAddr), amount, abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	PrintSuccess("Delegation transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Amount: %s wei", amount.String()))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	return nil
}

func createUndelegateTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	delegator, _ := cmd.Flags().GetString("delegator")
	validatorStr, _ := cmd.Flags().GetString("validator")
	amountStr, _ := cmd.Flags().GetString("amount")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(delegator, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse amount
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid amount: %s", amountStr))
		return
	}

	// Convert to wei
	weiAmount := new(big.Int).Mul(amount, big.NewInt(1e18))

	PrintInfo("Creating undelegation transaction")

	if err := innerCreateUndelegateTx(delegator, validatorStr, weiAmount, rpc); err != nil {
		PrintError("Failed to create undelegate transaction", err)
		return
	}
}

func innerCreateUndelegateTx(delegator, validator string, amount *big.Int, rpc string) error {
	outputFile := GeneratedFilePath(UndelegateFile)

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("undelegate", common.HexToAddress(validator), amount)
	if err != nil {
		return fmt.Errorf("failed to pack undelegate data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(delegator), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	PrintSuccess("Undelegation transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Amount: %s wei", amount.String()))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	PrintInfo("Unbonding period: 7 days")
	return nil
}

func createClaimRewardsTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	claimer, _ := cmd.Flags().GetString("claimer")
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(claimer, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating claim rewards transaction for delegator")

	if err := innerCreateClaimRewardsTx(claimer, validatorStr, rpc); err != nil {
		PrintError("Failed to create claim rewards transaction", err)
		return
	}
}

func createClaimValidatorRewardsTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating claim validator rewards transaction")

	if err := innerCreateClaimValidatorRewardsTx(validatorStr, rpc); err != nil {
		PrintError("Failed to create claim validator rewards transaction", err)
		return
	}
}

func innerCreateClaimRewardsTx(claimer, validator string, rpc string) error {
	outputFile := GeneratedFilePath(ClaimRewardsFile)

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("claimRewards", common.HexToAddress(validator))
	if err != nil {
		return fmt.Errorf("failed to pack claimRewards data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(claimer), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	PrintSuccess("Claim rewards transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	return nil
}

func innerCreateClaimValidatorRewardsTx(validator string, rpc string) error {
	outputFile := GeneratedFilePath("claimValidatorRewards.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	// Use the new claimValidatorRewards() method that doesn't require parameters
	abiData, err := stakingAbi.Pack("claimValidatorRewards")
	if err != nil {
		return fmt.Errorf("failed to pack claimValidatorRewards data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validator), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create raw transaction: %w", err)
	}

	PrintSuccess("Claim validator rewards transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	return nil
}

func queryDelegationInfo(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	delegatorStr, _ := cmd.Flags().GetString("delegator")
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(delegatorStr, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	delegator := common.HexToAddress(delegatorStr)
	validator := common.HexToAddress(validatorStr)

	_, stakingContract, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}
	delegationInfo, err := stakingContract.GetDelegationInfo(nil, delegator, validator)
	if err != nil {
		PrintError("Failed to get delegation info", err)
		return
	}

	// Convert wei to JU
	amount := delegationInfo.Amount
	pendingRewards := delegationInfo.PendingRewards

	amountJU := new(big.Int).Div(amount, big.NewInt(1e18))
	pendingRewardsJU := new(big.Int).Div(pendingRewards, big.NewInt(1e18))

	PrintSuccess("Delegation Information")
	fmt.Printf("Delegator: %s\n", delegatorStr)
	fmt.Printf("Validator: %s\n", validatorStr)
	fmt.Printf("Delegated Amount: %s JU\n", amountJU.String())
	fmt.Printf("Pending Rewards: %s JU\n", pendingRewardsJU.String())
}

// Query available unbonded amounts implementation
func queryAvailableUnbonded(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	delegatorStr, _ := cmd.Flags().GetString("delegator")
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(delegatorStr, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	delegator := common.HexToAddress(delegatorStr)
	validator := common.HexToAddress(validatorStr)

	_, stakingInstance, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}
	unbondingEntries, err := stakingInstance.GetUnbondingEntries(nil, delegator, validator)
	if err != nil {
		PrintError("Failed to get unbonding entries", err)
		return
	}
	PrintSuccess("Available Unbonded Information")
	for _, entry := range unbondingEntries {
		fmt.Printf("Amount: %s JU, Completion Block: %d\n", new(big.Int).Div(entry.Amount, big.NewInt(1e18)).String(), entry.CompletionBlock)
	}
}

// Implementation of the missing commands

// Unjail validator implementation
func createUnjailValidatorTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating unjail validator transaction")

	if err := innerCreateUnjailValidatorTx(validatorStr, rpc); err != nil {
		PrintError("Failed to create unjail validator transaction", err)
		return
	}
}

func innerCreateUnjailValidatorTx(validator string, rpc string) error {
	outputFile := GeneratedFilePath(UnjailValidatorFile)

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("unjailValidator", common.HexToAddress(validator))
	if err != nil {
		return fmt.Errorf("failed to pack unjailValidator data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validator), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create unjail validator transaction: %w", err)
	}

	PrintSuccess("Unjail validator transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	PrintWarning("Note: Validator must have passed a proposal before unjailing")
	return nil
}

func createIncreaseStakeTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")
	amountStr, _ := cmd.Flags().GetString("amount")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorAddr); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse amount
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid amount: %s", amountStr))
		return
	}

	// Convert to wei
	weiAmount := new(big.Int).Mul(amount, big.NewInt(1e18))

	PrintInfo("Creating increase stake transaction")

	if err := innerCreateIncreaseStakeTx(validatorAddr, weiAmount, rpc); err != nil {
		PrintError("Failed to create increase stake transaction", err)
		return
	}
}

func innerCreateIncreaseStakeTx(validatorAddr string, amount *big.Int, rpc string) error {
	outputFile := GeneratedFilePath("increaseStake.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	// Increase stake function name may vary depending on actual contract
	abiData, err := stakingAbi.Pack("addValidatorStake")
	if err != nil {
		return fmt.Errorf("failed to pack addValidatorStake data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validatorAddr), common.HexToAddress(StakingContractAddr), amount, abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create increase stake transaction: %w", err)
	}

	PrintSuccess("Increase stake transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Amount: %s wei", amount.String()))
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	return nil
}

func createDecreaseStakeTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")
	amountStr, _ := cmd.Flags().GetString("amount")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorAddr); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse amount
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid amount: %s", amountStr))
		return
	}

	// Convert to wei
	weiAmount := new(big.Int).Mul(amount, big.NewInt(1e18))

	PrintInfo("Creating decrease stake transaction")

	if err := innerCreateDecreaseStakeTx(validatorAddr, weiAmount, rpc); err != nil {
		PrintError("Failed to create decrease stake transaction", err)
		return
	}
}

func innerCreateDecreaseStakeTx(validatorAddr string, amount *big.Int, rpc string) error {
	outputFile := GeneratedFilePath("decreaseStake.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	// Decrease stake function name may vary depending on actual contract
	abiData, err := stakingAbi.Pack("decreaseValidatorStake", amount)
	if err != nil {
		return fmt.Errorf("failed to pack decreaseValidatorStake data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validatorAddr), common.HexToAddress(StakingContractAddr), nil, abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create decrease validator stake transaction: %w", err)
	}

	PrintSuccess("Decrease validator stake transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Amount: %s wei", amount.String()))
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	return nil
}

func createSetCommissionTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")
	rateStr, _ := cmd.Flags().GetString("rate")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorAddr); err != nil {
		PrintValidationError(err)
		return
	}

	// Parse commission rate
	rate, ok := new(big.Int).SetString(rateStr, 10)
	if !ok {
		PrintValidationError(fmt.Errorf("invalid commission rate: %s", rateStr))
		return
	}

	// Validate commission rate (0-10000)
	if rate.Cmp(big.NewInt(10000)) > 0 {
		PrintValidationError(fmt.Errorf("commission rate must be between 0 and 10000 (100%%)"))
		return
	}

	PrintInfo("Creating set commission rate transaction")

	if err := innerCreateSetCommissionTx(validatorAddr, rate, rpc); err != nil {
		PrintError("Failed to create set commission transaction", err)
		return
	}
}

func innerCreateSetCommissionTx(validatorAddr string, rate *big.Int, rpc string) error {
	outputFile := GeneratedFilePath("setCommission.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("updateCommissionRate", rate)
	if err != nil {
		return fmt.Errorf("failed to pack updateCommissionRate data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validatorAddr), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create set commission transaction: %w", err)
	}

	PrintSuccess("Set commission rate transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Commission rate: %s basis points", rate.String()))
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	return nil
}

func createDeregisterValidatorTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorAddr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating deregister validator transaction")

	if err := innerCreateDeregisterValidatorTx(validatorAddr, rpc); err != nil {
		PrintError("Failed to create deregister validator transaction", err)
		return
	}
}

func innerCreateDeregisterValidatorTx(validatorAddr string, rpc string) error {
	outputFile := GeneratedFilePath("resignValidator.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("resignValidator")
	if err != nil {
		return fmt.Errorf("failed to pack resignValidator data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validatorAddr), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create resign validator transaction: %w", err)
	}

	PrintSuccess("Resign validator transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	return nil
}

func createValidatorExitTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	validatorAddr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validatorAddr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating validator exit transaction")

	if err := innerCreateValidatorExitTx(validatorAddr, rpc); err != nil {
		PrintError("Failed to create validator exit transaction", err)
		return
	}
}

func innerCreateValidatorExitTx(validatorAddr string, rpc string) error {
	outputFile := GeneratedFilePath("exitValidator.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("exitValidator")
	if err != nil {
		return fmt.Errorf("failed to pack exitValidator data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(validatorAddr), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create validator exit transaction: %w", err)
	}

	PrintSuccess("Validator exit transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validatorAddr))
	return nil
}

func createWithdrawUnbondedTx(cmd *cobra.Command, args []string) {
	rpc := GetRPCEndpoint(cmd)
	claimer, _ := cmd.Flags().GetString("claimer")
	validatorStr, _ := cmd.Flags().GetString("validator")

	// Validate inputs
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(claimer, validatorStr); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Creating withdraw unbonded transaction")

	if err := innerCreateWithdrawUnbondedTx(claimer, validatorStr, rpc); err != nil {
		PrintError("Failed to create withdraw unbonded transaction", err)
		return
	}
}

func innerCreateWithdrawUnbondedTx(claimer, validator string, rpc string) error {
	outputFile := GeneratedFilePath("withdrawUnbonded.json")

	stakingAbi, err := abi.JSON(strings.NewReader(contracts.StakingABI))
	if err != nil {
		return fmt.Errorf("failed to parse staking ABI: %w", err)
	}

	abiData, err := stakingAbi.Pack("withdrawUnbonded", common.HexToAddress(validator), big.NewInt(10))
	if err != nil {
		return fmt.Errorf("failed to pack withdrawUnbonded data: %w", err)
	}

	err = CreateRawTx(common.HexToAddress(claimer), common.HexToAddress(StakingContractAddr), big.NewInt(0), abiData, rpc, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create withdraw unbonded transaction: %w", err)
	}

	PrintSuccess("Withdraw unbonded transaction created successfully!")
	PrintInfo(fmt.Sprintf("Transaction file: %s", outputFile))
	PrintInfo(fmt.Sprintf("Validator: %s", validator))
	return nil
}

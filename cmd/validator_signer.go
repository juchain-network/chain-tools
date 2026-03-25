package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func ValidatorSignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signer",
		Short: "Query the effective signer for a validator",
		Run:   queryValidatorSigner,
	}
	cmd.Flags().StringP("validator", "v", "", "Validator cold address")
	_ = cmd.MarkFlagRequired("validator")
	return cmd
}

func ValidatorBySignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-signer",
		Short: "Resolve the current validator by signer",
		Run:   queryValidatorBySigner,
	}
	cmd.Flags().String("signer", "", "Signer hot address")
	_ = cmd.MarkFlagRequired("signer")
	return cmd
}

func ValidatorSignerHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signer-history",
		Short: "Resolve a validator by historical signer binding",
		Run:   queryValidatorBySignerHistory,
	}
	cmd.Flags().String("signer", "", "Signer hot address")
	_ = cmd.MarkFlagRequired("signer")
	return cmd
}

func ValidatorPendingSignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-signer",
		Short: "Query the pending signer rotation for a validator",
		Run:   queryPendingSignerByValidator,
	}
	cmd.Flags().StringP("validator", "v", "", "Validator cold address")
	_ = cmd.MarkFlagRequired("validator")
	return cmd
}

func ValidatorPendingBySignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-by-signer",
		Short: "Resolve the validator that reserved a pending signer",
		Run:   queryPendingValidatorBySigner,
	}
	cmd.Flags().String("signer", "", "Pending signer hot address")
	_ = cmd.MarkFlagRequired("signer")
	return cmd
}

func ValidatorActiveSignersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-signers",
		Short: "List active runtime signers",
		Run:   queryActiveSigners,
	}
	return cmd
}

func ValidatorTopSignersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top-signers",
		Short: "List current effective top signers",
		Run:   queryTopSigners,
	}
	return cmd
}

func ValidatorEpochSignersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-signers",
		Short: "List checkpoint/header signers for epoch transition",
		Run:   queryEpochSigners,
	}
	return cmd
}

func ValidatorRewardSignersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward-signers",
		Short: "List reward-eligible signers with their validator stakes",
		Run:   queryRewardEligibleSigners,
	}
	return cmd
}

func queryValidatorSigner(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	validator, _ := cmd.Flags().GetString("validator")
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}
	if err := ValidateAddress(validator); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	signer, err := validatorInstance.GetValidatorSigner(&bind.CallOpts{}, common.HexToAddress(validator))
	if err != nil {
		PrintError("Failed to query validator signer", err)
		return
	}

	PrintSuccess("Validator signer")
	fmt.Printf("Validator: %s\n", validator)
	fmt.Printf("Effective Signer: %s\n", signer.Hex())
}

func queryValidatorBySigner(cmd *cobra.Command, _ []string) {
	querySignerLookup(cmd, false)
}

func queryValidatorBySignerHistory(cmd *cobra.Command, _ []string) {
	querySignerLookup(cmd, true)
}

func querySignerLookup(cmd *cobra.Command, historical bool) {
	rpc := GetRPCEndpoint(cmd)
	signer, _ := cmd.Flags().GetString("signer")
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}
	if err := ValidateAddress(signer); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	var validator common.Address
	if historical {
		validator, err = validatorInstance.GetValidatorBySignerHistory(&bind.CallOpts{}, common.HexToAddress(signer))
	} else {
		validator, err = validatorInstance.GetValidatorBySigner(&bind.CallOpts{}, common.HexToAddress(signer))
	}
	if err != nil {
		PrintError("Failed to resolve signer", err)
		return
	}

	PrintSuccess("Signer lookup")
	fmt.Printf("Signer: %s\n", signer)
	if historical {
		fmt.Printf("Historical Validator: %s\n", validator.Hex())
	} else {
		fmt.Printf("Current Validator: %s\n", validator.Hex())
	}
}

func queryPendingSignerByValidator(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	validator, _ := cmd.Flags().GetString("validator")
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}
	if err := ValidateAddress(validator); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}
	currentBlock, err := GetCurrentBlockNumber(rpc)
	if err != nil {
		PrintError("Failed to get current block number", err)
		return
	}

	pendingInfo, err := validatorInstance.GetPendingValidatorSigner(&bind.CallOpts{}, common.HexToAddress(validator))
	if err != nil {
		PrintError("Failed to query pending signer", err)
		return
	}

	PrintSuccess("Pending signer rotation")
	fmt.Printf("Validator: %s\n", validator)
	fmt.Printf("Pending Signer: %s\n", pendingInfo.Signer.Hex())
	fmt.Printf("Pending Effective Block: %s\n", pendingInfo.EffectiveBlock.String())
	fmt.Printf("Pending Recorded: %t\n", pendingInfo.Pending)
	fmt.Printf("Current Block: %d\n", currentBlock)
	fmt.Printf("Future Effective: %t\n", isFutureEffective(currentBlock, pendingInfo.EffectiveBlock, pendingInfo.Pending))
}

func queryPendingValidatorBySigner(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	signer, _ := cmd.Flags().GetString("signer")
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}
	if err := ValidateAddress(signer); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}
	currentBlock, err := GetCurrentBlockNumber(rpc)
	if err != nil {
		PrintError("Failed to get current block number", err)
		return
	}

	pendingInfo, err := validatorInstance.GetPendingValidatorBySigner(&bind.CallOpts{}, common.HexToAddress(signer))
	if err != nil {
		PrintError("Failed to query pending validator by signer", err)
		return
	}

	PrintSuccess("Pending signer reverse lookup")
	fmt.Printf("Signer: %s\n", signer)
	fmt.Printf("Validator: %s\n", pendingInfo.Validator.Hex())
	fmt.Printf("Pending Effective Block: %s\n", pendingInfo.EffectiveBlock.String())
	fmt.Printf("Pending Recorded: %t\n", pendingInfo.Pending)
	fmt.Printf("Current Block: %d\n", currentBlock)
	fmt.Printf("Future Effective: %t\n", isFutureEffective(currentBlock, pendingInfo.EffectiveBlock, pendingInfo.Pending))
}

func queryActiveSigners(cmd *cobra.Command, _ []string) {
	querySignerSet(cmd, "Active signers", func(v interface {
		GetActiveSigners(opts *bind.CallOpts) ([]common.Address, error)
	}) ([]common.Address, error) {
		return v.GetActiveSigners(&bind.CallOpts{})
	})
}

func queryTopSigners(cmd *cobra.Command, _ []string) {
	querySignerSet(cmd, "Top signers", func(v interface {
		GetTopSigners(opts *bind.CallOpts) ([]common.Address, error)
	}) ([]common.Address, error) {
		return v.GetTopSigners(&bind.CallOpts{})
	})
}

func queryEpochSigners(cmd *cobra.Command, _ []string) {
	querySignerSet(cmd, "Epoch transition signers", func(v interface {
		GetTopSignersForEpochTransition(opts *bind.CallOpts) ([]common.Address, error)
	}) ([]common.Address, error) {
		return v.GetTopSignersForEpochTransition(&bind.CallOpts{})
	})
}

func queryRewardEligibleSigners(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	result, err := validatorInstance.GetRewardEligibleSignersWithStakes(&bind.CallOpts{})
	if err != nil {
		PrintError("Failed to query reward-eligible signers", err)
		return
	}

	PrintSuccess(fmt.Sprintf("Reward-eligible signers (%d)", len(result.Signers)))
	for i, signer := range result.Signers {
		fmt.Printf("%d. Signer: %s\n", i+1, signer.Hex())
		if i < len(result.TotalStakes) {
			fmt.Printf("   Total Stake: %s\n", WeiToEther(result.TotalStakes[i]))
		}
	}
}

func querySignerSet[T any](cmd *cobra.Command, title string, fetch func(T) ([]common.Address, error)) {
	rpc := GetRPCEndpoint(cmd)
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	validatorInstance, _, _, err := GetContractInstance(rpc)
	if err != nil {
		PrintError("Failed to get contract instance", err)
		return
	}

	signers, err := fetch(any(validatorInstance).(T))
	if err != nil {
		PrintError(fmt.Sprintf("Failed to query %s", title), err)
		return
	}

	PrintSuccess(fmt.Sprintf("%s (%d)", title, len(signers)))
	for i, signer := range signers {
		fmt.Printf("%d. %s\n", i+1, signer.Hex())
	}
}

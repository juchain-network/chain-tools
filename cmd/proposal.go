package cmd

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"juchain.org/chain/tools/contracts"
)

func CreateProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a proposal",
		Run:   createProposalTx,
	}
	createProposalFlags(cmd)
	addOnlineSendFlags(cmd)
	return cmd
}

func createProposalFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposer", "p", "", "Validator cold address that creates the proposal")
	_ = cmd.MarkFlagRequired("proposer")
	cmd.Flags().StringP("target", "t", "", "Target validator cold address to add or remove")
	_ = cmd.MarkFlagRequired("target")
	cmd.Flags().StringP("operation", "o", "", "Operation type: add|remove")
	_ = cmd.MarkFlagRequired("operation")
}

func createProposalTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	proposer, _ := cmd.Flags().GetString("proposer")
	target, _ := cmd.Flags().GetString("target")
	operation, _ := cmd.Flags().GetString("operation")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(proposer, target); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateOperation(operation); err != nil {
		PrintValidationError(err)
		return
	}

	flag := operation == OperationAdd
	if flag {
		PrintInfo("Creating proposal to add new validator")
	} else {
		PrintInfo("Creating proposal to remove validator")
	}

	if err := innerCreateProposal(cmd, proposer, target, flag, rpc); err != nil {
		PrintError("Failed to create proposal", err)
		return
	}
}

func innerCreateProposal(cmd *cobra.Command, proposer, target string, flag bool, rpc string) error {
	proposalAbi, err := abi.JSON(strings.NewReader(contracts.ProposalABI))
	if err != nil {
		return fmt.Errorf("failed to parse proposal ABI: %w", err)
	}

	abiData, err := proposalAbi.Pack("createProposal", common.HexToAddress(target), flag, "")
	if err != nil {
		return fmt.Errorf("failed to pack proposal data: %w", err)
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(proposer),
		common.HexToAddress(ProposalContractAddr),
		nil,
		abiData,
		rpc,
		CreateProposalFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute proposal transaction: %w", err)
	}

	printTxExecutionResult(result, "Proposal transaction created successfully!")
	return nil
}

func CreateConfigProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create a config proposal",
		Run:   createConfigProposalTx,
	}
	createConfigProposalFlags(cmd)
	addOnlineSendFlags(cmd)
	return cmd
}

func createConfigProposalFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposer", "p", "", "Validator cold address that creates the config proposal")
	_ = cmd.MarkFlagRequired("proposer")
	cmd.Flags().Int64P("cid", "i", 0, "Config ID (0 proposalLastingPeriod, 1 punishThreshold, 2 removeThreshold, 3 decreaseRate, 4 withdrawProfitPeriod, 5 blockReward, 6 unbondingPeriod, 7 validatorUnjailPeriod, 8 minValidatorStake, 9 maxValidators, 10 minDelegation, 11 minUndelegation, 12 doubleSignSlashAmount, 13 doubleSignRewardAmount, 14 burnAddress, 15 doubleSignWindow, 16 commissionUpdateCooldown, 17 baseRewardRatio, 18 maxCommissionRate, 19 proposalCooldown)")
	_ = cmd.MarkFlagRequired("cid")
	cmd.Flags().StringP("value", "v", "", "New configuration value (decimal wei, 0x hex/address, or with unit: wei/gwei/ether/ju)")
	_ = cmd.MarkFlagRequired("value")
}

func createConfigProposalTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	proposer, _ := cmd.Flags().GetString("proposer")
	cid, _ := cmd.Flags().GetInt64("cid")
	rawValue, _ := cmd.Flags().GetString("value")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(proposer); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateConfigID(cid); err != nil {
		PrintValidationError(err)
		return
	}

	cvalue, err := ParseConfigValue(rawValue, cid)
	if err != nil {
		PrintValidationError(err)
		return
	}
	if cvalue.Sign() <= 0 {
		PrintValidationError(fmt.Errorf("config value must be positive: %s", rawValue))
		return
	}

	PrintInfo(fmt.Sprintf("Creating config update proposal for %s (ID: %d) with value: %s",
		GetConfigIDName(cid), cid, cvalue.String()))

	if err := innerCreateConfigProposal(cmd, proposer, cid, cvalue, rpc); err != nil {
		PrintError("Failed to create config proposal", err)
		return
	}
}

func innerCreateConfigProposal(cmd *cobra.Command, proposer string, cid int64, cvalue *big.Int, rpc string) error {
	proposalAbi, err := abi.JSON(strings.NewReader(contracts.ProposalABI))
	if err != nil {
		return fmt.Errorf("failed to parse proposal ABI: %w", err)
	}

	abiData, err := proposalAbi.Pack("createUpdateConfigProposal", big.NewInt(cid), cvalue)
	if err != nil {
		return fmt.Errorf("failed to pack config proposal data: %w", err)
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(proposer),
		common.HexToAddress(ProposalContractAddr),
		nil,
		abiData,
		rpc,
		CreateConfigProposalFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute config proposal transaction: %w", err)
	}

	printTxExecutionResult(result, "Config proposal transaction created successfully!")
	return nil
}

func VoteProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "Vote on a proposal",
		Run:   voteProposalTx,
	}
	voteProposalCmdFlags(cmd)
	addOnlineSendFlags(cmd)
	return cmd
}

func voteProposalCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("validator", "v", "", "Validator cold address that votes on the proposal")
	_ = cmd.MarkFlagRequired("validator")
	cmd.Flags().StringP("proposalId", "i", "", "Proposal ID (64-character hex string)")
	_ = cmd.MarkFlagRequired("proposalId")
	cmd.Flags().BoolP("approve", "a", false, "Approve this proposal (use -a for YES, omit for NO)")
}

func voteProposalTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	validator, _ := cmd.Flags().GetString("validator")
	proposalId, _ := cmd.Flags().GetString("proposalId")
	approve, _ := cmd.Flags().GetBool("approve")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddress(validator); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateProposalID(proposalId); err != nil {
		PrintValidationError(err)
		return
	}

	voteType := "REJECT"
	if approve {
		voteType = "APPROVE"
	}
	PrintInfo(fmt.Sprintf("Voting %s on proposal: %s", voteType, proposalId))

	if err := innerVoteProposal(cmd, validator, proposalId, approve, rpc); err != nil {
		PrintError("Failed to vote on proposal", err)
		return
	}
}

func innerVoteProposal(cmd *cobra.Command, validator, proposalId string, flag bool, rpc string) error {
	abiData, err := packVoteProposalData(proposalId, flag)
	if err != nil {
		return err
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(validator),
		common.HexToAddress(ProposalContractAddr),
		nil,
		abiData,
		rpc,
		VoteProposalFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute vote transaction: %w", err)
	}

	printTxExecutionResult(result, "Vote transaction created successfully!")
	return nil
}

func packVoteProposalData(proposalId string, flag bool) ([]byte, error) {
	proposalAbi, err := abi.JSON(strings.NewReader(contracts.ProposalABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse proposal ABI: %w", err)
	}

	var proposalIdBytes [32]byte
	copy(proposalIdBytes[:], common.HexToHash(proposalId).Bytes())

	abiData, err := proposalAbi.Pack("voteProposal", proposalIdBytes, flag)
	if err != nil {
		return nil, fmt.Errorf("failed to pack vote proposal data: %w", err)
	}
	return abiData, nil
}

// QueryProposalCmd creates a command to query a specific proposal
func QueryProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query a proposal by ID",
		Run:   queryProposalTx,
	}
	queryProposalFlags(cmd)
	return cmd
}

func queryProposalFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "Proposal ID (64-character hex string)")
	_ = cmd.MarkFlagRequired("id")
}

func queryProposalTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	proposalId, _ := cmd.Flags().GetString("id")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateProposalID(proposalId); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Fetching proposal details...")

	if err := innerQueryProposal(proposalId, rpc); err != nil {
		PrintError("Failed to query proposal", err)
		return
	}
}

func innerQueryProposal(proposalId string, rpc string) error {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	proposalContract, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		return fmt.Errorf("failed to instantiate proposal contract: %w", err)
	}

	var proposalIdBytes [32]byte
	copy(proposalIdBytes[:], common.HexToHash(proposalId).Bytes())

	proposal, err := proposalContract.Proposals(nil, proposalIdBytes)
	if err != nil {
		return fmt.Errorf("failed to query proposal: %w", err)
	}

	// Display proposal information
	fmt.Println("📋 Proposal Details:")
	fmt.Printf("Proposal ID: %s\n", proposalId)
	fmt.Printf("Proposer: %s (Validator Address)\n", proposal.Proposer.Hex())

	// Display different information based on proposal type
	if proposal.ProposalType.Int64() == 1 { // Validator management proposal
		if proposal.Flag {
			fmt.Printf("Target Address: %s (Validator to Add)\n", proposal.Dst.Hex())
			fmt.Printf("Action: Add New Validator (Flag: true)\n")
		} else {
			fmt.Printf("Target Address: %s (Validator to Remove)\n", proposal.Dst.Hex())
			fmt.Printf("Action: Remove Validator (Flag: false)\n")
		}
	} else if proposal.ProposalType.Int64() == 2 { // Configuration update proposal
		fmt.Printf("Config ID: %s (%s)\n", proposal.Cid.String(), getConfigIDName(proposal.Cid.Int64()))
		fmt.Printf("New Value: %s\n", proposal.NewValue.String())
		fmt.Printf("Action: Update Configuration\n")
	}

	fmt.Printf("Proposal Type: %s (%s)\n", proposal.ProposalType.String(), getProposalTypeName(proposal.ProposalType.Int64()))
	fmt.Printf("Create Time: %s\n", timeToString(proposal.CreateTime.Int64()))

	if proposal.Details != "" {
		fmt.Printf("Details: %s\n", proposal.Details)
	}

	return nil
}

// Get proposal type name
func getProposalTypeName(proposalType int64) string {
	switch proposalType {
	case 1:
		return "Validator Management"
	case 2:
		return "Configuration Update"
	default:
		return "Unknown Type"
	}
}

// Get configuration item name
func getConfigIDName(cid int64) string {
	switch cid {
	case 0:
		return "Proposal Lasting Period"
	case 1:
		return "Punish Threshold"
	case 2:
		return "Remove Threshold"
	case 3:
		return "Decrease Rate"
	case 4:
		return "Withdraw Profit Period"
	case 5:
		return "Block Reward"
	case 6:
		return "Unbonding Period"
	case 7:
		return "Validator Unjail Period"
	case 8:
		return "Min Validator Stake"
	case 9:
		return "Max Validators"
	case 10:
		return "Min Delegation"
	case 11:
		return "Min Undelegation"
	case 12:
		return "Double-Sign Slash Amount"
	case 13:
		return "Double-Sign Reward Amount"
	case 14:
		return "Burn Address"
	case 15:
		return "Double-Sign Window"
	case 16:
		return "Commission Update Cooldown"
	case 17:
		return "Base Reward Ratio"
	case 18:
		return "Max Commission Rate"
	case 19:
		return "Proposal Cooldown"
	default:
		return "Unknown Config"
	}
}

// QueryProposalsCmd creates a command to query all proposals
func QueryProposalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all proposals",
		Run:   queryProposalsTx,
	}
	return cmd
}

func queryProposalsTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo("Fetching all proposals...")

	if err := innerQueryProposals(rpc); err != nil {
		PrintError("Failed to query proposals", err)
		return
	}
}

func innerQueryProposals(rpc string) error {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	proposalContract, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		return fmt.Errorf("failed to instantiate proposal contract: %w", err)
	}

	// Get current block number
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Set query range (from genesis block to current block)
	opts := &bind.FilterOpts{
		Start: 0,
		End:   &currentBlock,
	}

	// Collect all proposal IDs
	proposalIDs := make(map[string]bool)

	// Query validator management proposal events
	validatorProposalIter, err := proposalContract.FilterLogCreateProposal(opts, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to filter validator proposals: %w", err)
	}
	defer validatorProposalIter.Close()

	for validatorProposalIter.Next() {
		event := validatorProposalIter.Event
		proposalID := common.BytesToHash(event.Id[:]).Hex()
		proposalIDs[proposalID] = true
	}

	// Query configuration update proposal events
	configProposalIter, err := proposalContract.FilterLogCreateConfigProposal(opts, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to filter config proposals: %w", err)
	}
	defer configProposalIter.Close()

	for configProposalIter.Next() {
		event := configProposalIter.Event
		proposalID := common.BytesToHash(event.Id[:]).Hex()
		proposalIDs[proposalID] = true
	}

	if len(proposalIDs) == 0 {
		fmt.Println("📋 No proposals found.")
		return nil
	}

	fmt.Printf("ℹ️  Found %d proposal(s):\n\n", len(proposalIDs))

	// Query detailed information for each proposal
	count := 1
	for proposalID := range proposalIDs {
		fmt.Printf("--- Proposal %d ---\n", count)

		var proposalIdBytes [32]byte
		copy(proposalIdBytes[:], common.HexToHash(proposalID).Bytes())

		proposal, err := proposalContract.Proposals(nil, proposalIdBytes)
		if err != nil {
			fmt.Printf("❌ Failed to query proposal %s: %v\n", proposalID, err)
			continue
		}

		// Display proposal information
		fmt.Printf("ID: %s\n", proposalID)
		fmt.Printf("Proposer: %s (Validator Address)\n", proposal.Proposer.Hex())

		// Display different information based on proposal type
		if proposal.ProposalType.Int64() == 1 { // Validator management proposal
			if proposal.Flag {
				fmt.Printf("Target: %s (Validator to Add)\n", proposal.Dst.Hex())
				fmt.Printf("Action: Add New Validator\n")
			} else {
				fmt.Printf("Target: %s (Validator to Remove)\n", proposal.Dst.Hex())
				fmt.Printf("Action: Remove Validator\n")
			}
		} else if proposal.ProposalType.Int64() == 2 { // Configuration update proposal
			fmt.Printf("Config ID: %s (%s)\n", proposal.Cid.String(), getConfigIDName(proposal.Cid.Int64()))
			fmt.Printf("New Value: %s\n", proposal.NewValue.String())
			fmt.Printf("Action: Update Configuration\n")
		}

		fmt.Printf("Type: %s (%s)\n", proposal.ProposalType.String(), getProposalTypeName(proposal.ProposalType.Int64()))
		fmt.Printf("Create Time: %s\n", timeToString(proposal.CreateTime.Int64()))

		// Query proposal voting results and status
		result, err := proposalContract.Results(nil, proposalIdBytes)
		if err != nil {
			fmt.Printf("⚠️  Status: Cannot query result (%v)\n", err)
		} else {
			statusIcon, statusText := getProposalStatus(result.Agree, result.Reject, result.ResultExist)
			fmt.Printf("Status: %s %s\n", statusIcon, statusText)
			if result.Agree > 0 || result.Reject > 0 {
				fmt.Printf("Votes: 👍 %d agree, 👎 %d reject\n", result.Agree, result.Reject)
			}
		}

		if proposal.Details != "" {
			fmt.Printf("Details: %s\n", proposal.Details)
		}

		fmt.Println()
		count++
	}

	return nil
}

// Time conversion helper function
func timeToString(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}
	return time.Unix(timestamp, 0).UTC().String()
}

// Helper function to get proposal status
func getProposalStatus(agree uint16, reject uint16, resultExist bool) (string, string) {
	if !resultExist {
		if agree == 0 && reject == 0 {
			return "⏳", "Pending (No votes yet)"
		} else {
			return "⏳", "Pending (Voting in progress)"
		}
	}

	// resultExist = true means the proposal has a final result
	// According to Proposal.sol logic: pass if more than half agree, fail if more than half reject
	if agree > reject {
		return "✅", "Passed"
	} else {
		return "❌", "Rejected"
	}
}

// QueryParamCmd creates a command to query a specific governance parameter
func QueryParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "param",
		Short: "Query a governance parameter value",
		Run:   queryParamTx,
	}
	queryParamFlags(cmd)
	return cmd
}

func queryParamFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("cid", "c", -1, "Config ID (0-19)")
	_ = cmd.MarkFlagRequired("cid")
}

func queryParamTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	cid, _ := cmd.Flags().GetInt64("cid")

	// Validate input parameters
	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateConfigID(cid); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo(fmt.Sprintf("Fetching value for config ID %d (%s)...", cid, GetConfigIDName(cid)))

	if err := innerQueryParam(cid, rpc); err != nil {
		PrintError("Failed to query parameter", err)
		return
	}
}

func innerQueryParam(cid int64, rpc string) error {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	proposalContract, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		return fmt.Errorf("failed to instantiate proposal contract: %w", err)
	}

	var value interface{}
	var errQuery error

	opts := &bind.CallOpts{Context: context.Background()}

	switch cid {
	case 0:
		value, errQuery = proposalContract.ProposalLastingPeriod(opts)
	case 1:
		value, errQuery = proposalContract.PunishThreshold(opts)
	case 2:
		value, errQuery = proposalContract.RemoveThreshold(opts)
	case 3:
		value, errQuery = proposalContract.DecreaseRate(opts)
	case 4:
		value, errQuery = proposalContract.WithdrawProfitPeriod(opts)
	case 5:
		value, errQuery = proposalContract.BlockReward(opts)
	case 6:
		value, errQuery = proposalContract.UnbondingPeriod(opts)
	case 7:
		value, errQuery = proposalContract.ValidatorUnjailPeriod(opts)
	case 8:
		value, errQuery = proposalContract.MinValidatorStake(opts)
	case 9:
		value, errQuery = proposalContract.MaxValidators(opts)
	case 10:
		value, errQuery = proposalContract.MinDelegation(opts)
	case 11:
		value, errQuery = proposalContract.MinUndelegation(opts)
	case 12:
		value, errQuery = proposalContract.DoubleSignSlashAmount(opts)
	case 13:
		value, errQuery = proposalContract.DoubleSignRewardAmount(opts)
	case 14:
		value, errQuery = proposalContract.BurnAddress(opts)
	case 15:
		value, errQuery = proposalContract.DoubleSignWindow(opts)
	case 16:
		value, errQuery = proposalContract.CommissionUpdateCooldown(opts)
	case 17:
		value, errQuery = proposalContract.BaseRewardRatio(opts)
	case 18:
		value, errQuery = proposalContract.MaxCommissionRate(opts)
	case 19:
		value, errQuery = proposalContract.ProposalCooldown(opts)
	default:
		return fmt.Errorf("unknown config ID: %d", cid)
	}

	if errQuery != nil {
		return fmt.Errorf("failed to query contract: %w", errQuery)
	}

	fmt.Println("📋 Parameter Details:")
	fmt.Printf("ID: %d\n", cid)
	fmt.Printf("Name: %s\n", GetConfigIDName(cid))

	if cid == 14 {
		// Address type
		fmt.Printf("Value: %s\n", value.(common.Address).Hex())
	} else {
		// BigInt type
		fmt.Printf("Value: %s\n", value.(*big.Int).String())
	}

	return nil
}

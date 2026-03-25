package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "JuChain chain management and governance CLI",
	Long: `cli is a command line tool for JuChain chain management and governance.
It supports the signer-separated validator model:
- validator: cold address for governance and staking ownership
- signer: hot address for block production
- feeAddr: reward receiving address

Features:
- Create and vote on validator addition/removal proposals
- Create and vote on configuration update proposals
- Query validator/signer state and manage rewards
- Submit double-sign evidence
- Generate transactions for offline signing or send them directly online

Use "cli [command] --help" for more information about a command.`,
	PersistentPreRun: validateGlobalFlags,
}

func validateGlobalFlags(cmd *cobra.Command, args []string) {
	// Skip validation for help and version commands
	if cmd.Name() == "help" || cmd.Name() == "version" {
		return
	}

	// Check if command requires RPC connection
	requiresRPC := map[string]struct{}{
		"list":                    {},
		"query":                   {},
		"create":                  {},
		"config":                  {},
		"vote":                    {},
		"param":                   {},
		"send":                    {},
		"transfer":                {},
		"edit":                    {},
		"claim":                   {},
		"signer":                  {},
		"by-signer":               {},
		"signer-history":          {},
		"pending-signer":          {},
		"pending-by-signer":       {},
		"active-signers":          {},
		"top-signers":             {},
		"epoch-signers":           {},
		"reward-signers":          {},
		"submit-double-sign":      {},
		"validator-register":      {},
		"delegate":                {},
		"undelegate":              {},
		"claim-rewards":           {},
		"claim-validator-rewards": {},
		"query-delegation":        {},
		"query-unbonded":          {},
		"stake-increase":          {},
		"stake-decrease":          {},
		"set-commission":          {},
		"validator-deregister":    {},
		"validator-exit":          {},
		"validator-unjail":        {},
		"withdraw-unbonded":       {},
	}
	cmdName := cmd.Name()

	if _, ok := requiresRPC[cmdName]; ok {
		rpc := GetRPCEndpoint(cmd) // Use config-aware function instead of flag only
		if rpc == "" {
			PrintValidationError(fmt.Errorf("RPC endpoint is required for command '%s'", cmdName))
			os.Exit(1)
		}
		if err := ValidateRPCURL(rpc); err != nil {
			PrintValidationError(err)
			os.Exit(1)
		}
	}
}

// GetRPCEndpoint gets RPC endpoint from global flag
func GetRPCEndpoint(cmd *cobra.Command) string {
	return rpcEndpoint
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		PrintError("Command execution failed", err)
		os.Exit(1)
	}
}

// Parent command declarations
var (
	// proposalCmd - For all proposal-related operations
	proposalCmd = &cobra.Command{
		Use:   "proposal",
		Short: "Manage proposals on the blockchain",
		Long:  `Create, vote, and query proposals on the Juchain blockchain.`,
	}

	// validatorCmd - For validator management
	validatorCmd = &cobra.Command{
		Use:   "validator",
		Short: "Manage validators on the blockchain",
		Long:  `List validators, inspect signer state, edit fee/signer metadata, and manage validator rewards.`,
	}

	// stakingCmd - For staking operations
	stakingCmd = &cobra.Command{
		Use:   "staking",
		Short: "Manage staking operations",
		Long:  `Delegate, undelegate, and claim rewards from staking on the Juchain blockchain.`,
	}

	// punishCmd - For punishment and evidence submission
	punishCmd = &cobra.Command{
		Use:   "punish",
		Short: "Manage punish and evidence operations",
		Long:  `Submit punish-related evidence such as double-sign reports on the Juchain blockchain.`,
	}

	// miscCmd - For miscellaneous commands
	miscCmd = &cobra.Command{
		Use:   "misc",
		Short: "Miscellaneous commands",
		Long:  `Offline transaction signing, broadcast, and other miscellaneous commands.`,
	}
)

var (
	// Global flags
	rpcEndpoint string
)

func init() {
	// Add global flags to root command
	rootCmd.PersistentFlags().StringVar(&rpcEndpoint, "rpc", "", "RPC endpoint URL")

	// Group commands under parent commands

	// Proposal commands
	proposalCmd.AddCommand(
		CreateProposalCmd(),
		CreateConfigProposalCmd(),
		VoteProposalCmd(),
		QueryProposalCmd(),
		QueryProposalsCmd(),
		QueryParamCmd(),
	)

	// Validator commands - validator management
	validatorCmd.AddCommand(
		ValidatorsCmd(), // list validators
		ValidatorCmd(),  // query validator info
		ValidatorSignerCmd(),
		ValidatorBySignerCmd(),
		ValidatorSignerHistoryCmd(),
		ValidatorPendingSignerCmd(),
		ValidatorPendingBySignerCmd(),
		ValidatorActiveSignersCmd(),
		ValidatorTopSignersCmd(),
		ValidatorEpochSignersCmd(),
		ValidatorRewardSignersCmd(),
		EditValidatorCmd(),   // edit validator information
		WithdrawProfitsCmd(), // withdraw validator profits
	)

	// Staking commands - staking and delegation operations
	stakingCmd.AddCommand(
		RegisterValidatorCmd(),      // register validator staking
		DelegateCmd(),               // delegate to validator
		UndelegateCmd(),             // undelegate from validator
		IncreaseStakeCmd(),          // increase validator stake
		DecreaseStakeCmd(),          // decrease validator stake
		SetCommissionCmd(),          // set validator commission rate
		DeregisterValidatorCmd(),    // validator deregistration
		ValidatorExitCmd(),          // validator complete exit
		UnjailValidatorCmd(),        // unjail a validator
		ClaimRewardsCmd(),           // claim staking rewards (delegators)
		ClaimValidatorRewardsCmd(),  // claim validator rewards (validators)
		WithdrawUnbondedCmd(),       // withdraw unbonded stakes
		QueryDelegationCmd(),        // query delegation information
		QueryAvailableUnbondedCmd(), // query available unbonded amounts
	)

	punishCmd.AddCommand(
		SubmitDoubleSignCmd(),
	)

	// Misc commands
	miscCmd.AddCommand(
		TransferCmd(),
		SignRawTxCmd(),
		SendSignedTxCmd(),
	)

	// Add parent commands to root
	rootCmd.AddCommand(
		proposalCmd,
		validatorCmd,
		stakingCmd,
		punishCmd,
		miscCmd,
	)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ju-cli",
	Short: "JuChain blockchain governance command line tool",
	Long: `JuChain CLI is a command line tool for JuChain blockchain governance.
It provides comprehensive functionality for validator management and proposal voting.

Features:
- Create and vote on validator addition/removal proposals
- Create and vote on configuration update proposals  
- Query validator information and manage rewards
- Sign and broadcast transactions to the blockchain network

Use "ju-cli [command] --help" for more information about a command.`,
	PersistentPreRun: validateGlobalFlags,
}

func validateGlobalFlags(cmd *cobra.Command, args []string) {
	// Skip validation for help and version commands
	if cmd.Name() == "help" || cmd.Name() == "version" {
		return
	}

	// Check if command requires RPC connection
	requiresRPC := []string{
		"list", "query", "create", "config", "vote", "withdraw_profits", "param",
		"send", "register-validator", "edit-validator", "delegate", "undelegate",
		"claim-rewards", "claim-validator-rewards", "query-validator", "query-delegation", "list-top-validators",
		"increase-stake", "decrease-stake", "set-commission", "deregister", "exit",
		"claim-unbonding-rewards",
	}
	cmdName := cmd.Name()

	for _, name := range requiresRPC {
		if cmdName == name {
			rpc := GetRPCEndpoint(cmd) // Use config-aware function instead of flag only
			if rpc == "" {
				PrintValidationError(fmt.Errorf("RPC endpoint is required for command '%s'", cmdName))
				os.Exit(1)
			}
			if err := ValidateRPCURL(rpc); err != nil {
				PrintValidationError(err)
				os.Exit(1)
			}
			break
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
		Long:  `List, register, edit, and query validators on the Juchain blockchain.`,
	}

	// stakingCmd - For staking operations
	stakingCmd = &cobra.Command{
		Use:   "staking",
		Short: "Manage staking operations",
		Long:  `Delegate, undelegate, and claim rewards from staking on the Juchain blockchain.`,
	}

	// miscCmd - For miscellaneous commands
	miscCmd = &cobra.Command{
		Use:   "misc",
		Short: "Miscellaneous commands",
		Long:  `Configuration, transaction signing, and other miscellaneous commands.`,
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
		ValidatorsCmd(),      // list validators
		ValidatorCmd(),       // query validator info
		EditValidatorCmd(),   // edit validator information
		WithdrawProfitsCmd(), // withdraw validator profits
	)

	// Staking commands - staking and delegation operations
	stakingCmd.AddCommand(
		RegisterValidatorCmd(),          // register validator staking
		DelegateCmd(),                   // delegate to validator
		UndelegateCmd(),                 // undelegate from validator
		IncreaseStakeCmd(),              // increase validator stake
		DecreaseStakeCmd(),              // decrease validator stake
		SetCommissionCmd(),              // set validator commission rate
		DeregisterValidatorCmd(),        // validator deregistration
		ValidatorExitCmd(),              // validator complete exit
		UnjailValidatorCmd(),            // unjail a validator
		ClaimRewardsCmd(),               // claim staking rewards (delegators)
		ClaimValidatorRewardsCmd(),      // claim validator rewards (validators)
		WithdrawUnbondedCmd(),           // withdraw unbonded stakes
		QueryDelegationCmd(),            // query delegation information
		QueryAvailableUnbondedCmd(),     // query available unbonded amounts
	)

	// Misc commands
	miscCmd.AddCommand(
		SignRawTxCmd(),
		SendSignedTxCmd(),
	)

	// Add parent commands to root
	rootCmd.AddCommand(
		proposalCmd,
		validatorCmd,
		stakingCmd,
		miscCmd,
	)
}

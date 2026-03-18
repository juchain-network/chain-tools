package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func SendSignedTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send signed transaction to remote RPC endpoint",
		Run:   sendSignedTx,
	}
	sendSignedTxCmdFlags(cmd)
	return cmd
}

func sendSignedTxCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Signed transaction file")
	_ = cmd.MarkFlagRequired("file")
}

func sendSignedTx(cmd *cobra.Command, _ []string) {
	file, _ := cmd.Flags().GetString("file")
	rpc := GetRPCEndpoint(cmd) // Use config-aware function
	file = ResolveExistingFile(file)

	// Validate input parameters
	if err := ValidateFile(file); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo(fmt.Sprintf("Broadcasting signed transaction from: %s", file))
	if err := innerSendSignedTx(file, rpc); err != nil {
		PrintError("Failed to broadcast transaction", err)
		return
	}
}

func innerSendSignedTx(file, rpc string) error {
	file = ResolveExistingFile(file)
	txHash, err := SendSignedTx(rpc, file)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	PrintSuccess("Transaction broadcast successfully!")
	PrintInfo(fmt.Sprintf("Transaction hash: %s", txHash.Hex()))
	return nil
}

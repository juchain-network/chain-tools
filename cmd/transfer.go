package cmd

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a native ETH transfer transaction",
		Long:  "Create a native ETH transfer transaction to send funds to another address",
		Run:   createTransferTx,
	}

	cmd.Flags().StringP("from", "f", "", "Sender address (required)")
	cmd.Flags().StringP("to", "t", "", "Recipient address (required)")
	cmd.Flags().StringP("amount", "a", "", "Transfer amount in ETH (supports decimals, required)")

	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("amount")
	addOnlineSendFlags(cmd)

	return cmd
}

func createTransferTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	rawAmount, _ := cmd.Flags().GetString("amount")

	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}

	if err := ValidateAddresses(from, to); err != nil {
		PrintValidationError(err)
		return
	}

	amountWei, err := ParseTransferAmount(rawAmount)
	if err != nil {
		PrintValidationError(err)
		return
	}
	if amountWei.Sign() <= 0 {
		PrintValidationError(fmt.Errorf("transfer amount must be positive"))
		return
	}

	PrintInfo(fmt.Sprintf(
		"Creating transfer of %s ETH (%s wei) to %s",
		strings.TrimSpace(rawAmount),
		amountWei.String(),
		to,
	))

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(from),
		common.HexToAddress(to),
		amountWei,
		nil,
		rpc,
		TransferFile,
	)
	if err != nil {
		PrintError("Failed to create transfer transaction", err)
		return
	}

	printTxExecutionResult(result, "Transfer transaction created successfully!")
}

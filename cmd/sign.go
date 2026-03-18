package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func SignRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign raw transaction from file",
		Run:   signRawTx,
	}
	signRawTxCmdFlags(cmd)
	return cmd
}

func signRawTxCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Raw transaction file")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringP("wallet", "w", "", "Signer wallet file")
	cmd.Flags().StringP("private-key", "k", "", "Signer private key (hex string)")
	cmd.Flags().StringP("password", "p", "", "Signer password file (required when using wallet file)")
}

func signRawTx(cmd *cobra.Command, _ []string) {
	file, _ := cmd.Flags().GetString("file")
	wallet, _ := cmd.Flags().GetString("wallet")
	privateKeyStr, _ := cmd.Flags().GetString("private-key")
	passwordFile, _ := cmd.Flags().GetString("password")
	file = ResolveExistingFile(file)

	// Validate input parameters
	if err := ValidateFile(file); err != nil {
		PrintValidationError(err)
		return
	}

	var privateKey *ecdsa.PrivateKey
	var err error
	signOpts := signerOptions{
		Wallet:     wallet,
		PrivateKey: privateKeyStr,
		Password:   passwordFile,
	}
	privateKey, err = loadSignerPrivateKey(signOpts)
	if err != nil {
		PrintValidationError(err)
		return
	}

	if wallet != "" {
		PrintInfo(fmt.Sprintf("Signing transaction from file: %s using wallet: %s", file, wallet))
	} else {
		PrintInfo(fmt.Sprintf("Signing transaction from file: %s using private key", file))
	}

	if err := innerSignRawTx(file, privateKey); err != nil {
		PrintError("Failed to sign transaction", err)
		return
	}
}

func fethchKeyFromFile(path string) (string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines[0], nil
}

func innerSignRawTx(file string, privateKey *ecdsa.PrivateKey) error {
	file = ResolveExistingFile(file)
	outputFile := addSuffixToFilename(file, "_signed")

	err := SignRawTx(file, privateKey, outputFile)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	PrintSuccess("Transaction signed successfully!")
	PrintInfo(fmt.Sprintf("Signed transaction saved to: %s", outputFile))
	return nil
}

func addSuffixToFilename(filename, suffix string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return base + suffix + ext
}

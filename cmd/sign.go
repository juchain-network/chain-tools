package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
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

	// Check that either wallet or private-key is provided, but not both
	walletProvided := wallet != ""
	privateKeyProvided := privateKeyStr != ""

	if !walletProvided && !privateKeyProvided {
		PrintValidationError(fmt.Errorf("must provide either --wallet or --private-key"))
		return
	}

	if walletProvided && privateKeyProvided {
		PrintValidationError(fmt.Errorf("cannot provide both --wallet and --private-key"))
		return
	}

	var privateKey *ecdsa.PrivateKey
	var err error

	if walletProvided {
		// Validate wallet file
		if err := ValidateFile(wallet); err != nil {
			PrintValidationError(err)
			return
		}

		// Validate password file (required for wallet)
		if passwordFile == "" {
			PrintValidationError(fmt.Errorf("--password is required when using --wallet"))
			return
		}

		if err := ValidateFile(passwordFile); err != nil {
			PrintValidationError(err)
			return
		}

		PrintInfo(fmt.Sprintf("Signing transaction from file: %s using wallet: %s", file, wallet))

		// Read password
		password, err := fethchKeyFromFile(passwordFile)
		if err != nil {
			PrintError("Failed to read password file", err)
			return
		}

		// Decrypt wallet file
		privateKey, err = ReadKeystoreFile(wallet, password)
		if err != nil {
			PrintError("Failed to decrypt keystore file", err)
			return
		}
	} else {
		// Using private key string
		PrintInfo(fmt.Sprintf("Signing transaction from file: %s using private key", file))

		// Parse private key string
		privateKeyStr = strings.TrimPrefix(privateKeyStr, "0x")
		privateKey, err = crypto.HexToECDSA(privateKeyStr)
		if err != nil {
			PrintError("Failed to parse private key", err)
			return
		}
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

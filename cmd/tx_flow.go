package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

type signerOptions struct {
	Wallet     string
	PrivateKey string
	Password   string
}

type txExecutionOptions struct {
	Send   bool
	Signer signerOptions
}

type txExecutionResult struct {
	Online     bool
	OutputFile string
	TxHash     common.Hash
}

func addOnlineSendFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("send", false, "Sign and broadcast immediately instead of writing a raw transaction file")
	cmd.Flags().String("wallet", "", "Keystore wallet file used for online send")
	cmd.Flags().String("private-key", "", "Private key used for online send (hex string)")
	cmd.Flags().String("password", "", "Password file for --wallet during online send")
}

func getTxExecutionOptions(cmd *cobra.Command) (txExecutionOptions, error) {
	if cmd == nil {
		return txExecutionOptions{}, nil
	}

	send, _ := cmd.Flags().GetBool("send")
	opts := txExecutionOptions{
		Send: send,
		Signer: signerOptions{
			Wallet:     mustGetStringFlag(cmd, "wallet"),
			PrivateKey: mustGetStringFlag(cmd, "private-key"),
			Password:   mustGetStringFlag(cmd, "password"),
		},
	}

	hasSignerFlags := opts.Signer.Wallet != "" || opts.Signer.PrivateKey != "" || opts.Signer.Password != ""
	if !opts.Send {
		if hasSignerFlags {
			return txExecutionOptions{}, fmt.Errorf("--wallet, --private-key, and --password require --send")
		}
		return opts, nil
	}

	if err := validateSignerOptions(opts.Signer); err != nil {
		return txExecutionOptions{}, err
	}
	return opts, nil
}

func validateSignerOptions(opts signerOptions) error {
	walletProvided := opts.Wallet != ""
	privateKeyProvided := opts.PrivateKey != ""

	if !walletProvided && !privateKeyProvided {
		return fmt.Errorf("must provide either --wallet or --private-key when using --send")
	}
	if walletProvided && privateKeyProvided {
		return fmt.Errorf("cannot provide both --wallet and --private-key")
	}

	if walletProvided {
		if err := ValidateFile(opts.Wallet); err != nil {
			return err
		}
		if opts.Password == "" {
			return fmt.Errorf("--password is required when using --wallet")
		}
		if err := ValidateFile(opts.Password); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func loadSignerPrivateKey(opts signerOptions) (*ecdsa.PrivateKey, error) {
	if err := validateSignerOptions(opts); err != nil {
		return nil, err
	}

	if opts.Wallet != "" {
		password, err := fethchKeyFromFile(opts.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to read password file: %w", err)
		}
		privateKey, err := ReadKeystoreFile(opts.Wallet, password)
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(opts.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return privateKey, nil
}

func validateSignerMatchesSender(privateKey *ecdsa.PrivateKey, sender common.Address) error {
	signerAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	if signerAddr != sender {
		return fmt.Errorf("signer address mismatch: expected %s, got %s", sender.Hex(), signerAddr.Hex())
	}
	return nil
}

func executeTransaction(cmd *cobra.Command, sender, contract common.Address, value *big.Int, data []byte, rpc, outputFile string) (*txExecutionResult, error) {
	options, err := getTxExecutionOptions(cmd)
	if err != nil {
		return nil, err
	}

	if options.Send {
		privateKey, err := loadSignerPrivateKey(options.Signer)
		if err != nil {
			return nil, err
		}
		if err := validateSignerMatchesSender(privateKey, sender); err != nil {
			return nil, err
		}

		txHash, err := CreateAndSendTx(sender, contract, value, data, rpc, privateKey)
		if err != nil {
			return nil, err
		}
		return &txExecutionResult{
			Online: true,
			TxHash: txHash,
		}, nil
	}

	outputFile = GeneratedFilePath(outputFile)
	if err := CreateRawTx(sender, contract, value, data, rpc, outputFile); err != nil {
		return nil, err
	}
	return &txExecutionResult{
		OutputFile: outputFile,
	}, nil
}

func mustGetStringFlag(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}

func printTxExecutionResult(result *txExecutionResult, successMessage string) {
	PrintSuccess(successMessage)
	if result.Online {
		PrintInfo(fmt.Sprintf("Transaction hash: %s", result.TxHash.Hex()))
		return
	}
	PrintInfo(fmt.Sprintf("Transaction file: %s", result.OutputFile))
}

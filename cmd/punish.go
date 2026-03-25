package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"juchain.org/chain/tools/contracts"
)

var hexBytesRegexp = regexp.MustCompile(`^(0[xX])?[0-9a-fA-F]+$`)

func SubmitDoubleSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-double-sign",
		Short: "Submit double-sign evidence",
		Long:  "Submit two conflicting RLP-encoded block headers as double-sign evidence",
		Run:   submitDoubleSignEvidenceTx,
	}

	cmd.Flags().StringP("reporter", "r", "", "Reporter address that submits the evidence")
	cmd.Flags().String("header1", "", "Hex-encoded RLP block header #1")
	cmd.Flags().String("header2", "", "Hex-encoded RLP block header #2")
	cmd.Flags().String("header1-file", "", "File containing block header #1 (raw bytes or hex text)")
	cmd.Flags().String("header2-file", "", "File containing block header #2 (raw bytes or hex text)")
	_ = cmd.MarkFlagRequired("reporter")
	addOnlineSendFlags(cmd)

	return cmd
}

func submitDoubleSignEvidenceTx(cmd *cobra.Command, _ []string) {
	rpc := GetRPCEndpoint(cmd)
	reporter, _ := cmd.Flags().GetString("reporter")
	header1Hex, _ := cmd.Flags().GetString("header1")
	header2Hex, _ := cmd.Flags().GetString("header2")
	header1File, _ := cmd.Flags().GetString("header1-file")
	header2File, _ := cmd.Flags().GetString("header2-file")

	if err := ValidateRPCURL(rpc); err != nil {
		PrintValidationError(err)
		return
	}
	if err := ValidateAddress(reporter); err != nil {
		PrintValidationError(err)
		return
	}

	header1, err := loadDoubleSignHeader(header1Hex, header1File, "header1")
	if err != nil {
		PrintValidationError(err)
		return
	}
	header2, err := loadDoubleSignHeader(header2Hex, header2File, "header2")
	if err != nil {
		PrintValidationError(err)
		return
	}

	PrintInfo(fmt.Sprintf(
		"Submitting double-sign evidence: reporter=%s, header1=%d bytes, header2=%d bytes",
		reporter,
		len(header1),
		len(header2),
	))

	if err := innerSubmitDoubleSignEvidence(cmd, reporter, header1, header2, rpc); err != nil {
		PrintError("Failed to submit double-sign evidence", err)
		return
	}
}

func innerSubmitDoubleSignEvidence(cmd *cobra.Command, reporter string, header1, header2 []byte, rpc string) error {
	abiData, err := packSubmitDoubleSignEvidenceData(header1, header2)
	if err != nil {
		return err
	}

	result, err := executeTransaction(
		cmd,
		common.HexToAddress(reporter),
		common.HexToAddress(PunishContractAddr),
		nil,
		abiData,
		rpc,
		SubmitDoubleSignEvidenceFile,
	)
	if err != nil {
		return fmt.Errorf("failed to execute double-sign evidence transaction: %w", err)
	}

	printTxExecutionResult(result, "Double-sign evidence transaction created successfully!")
	return nil
}

func packSubmitDoubleSignEvidenceData(header1, header2 []byte) ([]byte, error) {
	punishAbi, err := abi.JSON(strings.NewReader(contracts.PunishABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse punish ABI: %w", err)
	}

	abiData, err := punishAbi.Pack("submitDoubleSignEvidence", header1, header2)
	if err != nil {
		return nil, fmt.Errorf("failed to pack double-sign evidence data: %w", err)
	}
	return abiData, nil
}

func loadDoubleSignHeader(hexInput, fileInput, field string) ([]byte, error) {
	if strings.TrimSpace(hexInput) == "" && strings.TrimSpace(fileInput) == "" {
		return nil, fmt.Errorf("%s is required via --%s or --%s-file", field, field, field)
	}
	if strings.TrimSpace(hexInput) != "" && strings.TrimSpace(fileInput) != "" {
		return nil, fmt.Errorf("%s accepts either --%s or --%s-file, not both", field, field, field)
	}

	if strings.TrimSpace(hexInput) != "" {
		return parseHexHeader(hexInput, field)
	}

	filePath := ResolveExistingFile(strings.TrimSpace(fileInput))
	if err := ValidateFile(filePath); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", field, err)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed != "" && hexBytesRegexp.MatchString(trimmed) {
		return parseHexHeader(trimmed, field)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("%s file is empty", field)
	}
	return content, nil
}

func parseHexHeader(raw, field string) ([]byte, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")
	if value == "" {
		return nil, fmt.Errorf("%s cannot be empty", field)
	}
	if len(value)%2 != 0 {
		return nil, fmt.Errorf("%s hex input must have even length", field)
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%s must be valid hex: %w", field, err)
	}
	if len(decoded) == 0 {
		return nil, fmt.Errorf("%s cannot decode to empty bytes", field)
	}
	return decoded, nil
}

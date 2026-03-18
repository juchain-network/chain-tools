package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"juchain.org/chain/tools/contracts"
)

func buildUnsignedTx(client *ethclient.Client, caller, contract common.Address, value *big.Int, data []byte) (*types.Transaction, *big.Int, error) {
	nonce, err := client.PendingNonceAt(context.Background(), caller)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	msg := ethereum.CallMsg{
		From:  caller,
		To:    &contract,
		Data:  data,
		Value: value,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, nil, fmt.Errorf("gas estimation failed: %v", err)
	}
	gasLimit = gasLimit * DefaultGasMultiplier / 100

	tx := types.NewTransaction(
		nonce,
		contract,
		value,
		gasLimit,
		gasPrice,
		data,
	)
	return tx, chainID, nil
}

func CreateRawTx(
	caller common.Address,
	contract common.Address,
	value *big.Int,
	data []byte,
	rpc string,
	output string,
) error {
	output = GeneratedFilePath(output)
	if err := EnsureParentDir(output); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	client, err := ethclient.Dial(rpc)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %v", err)
	}
	defer client.Close()

	tx, chainID, err := buildUnsignedTx(client, caller, contract, value, data)
	if err != nil {
		return err
	}

	rawTx := map[string]interface{}{
		"nonce":    tx.Nonce(),
		"gasPrice": tx.GasPrice(),
		"gasLimit": tx.Gas(),
		"to":       tx.To().Hex(),
		"value":    tx.Value(),
		"data":     common.Bytes2Hex(tx.Data()),
		"chainId":  chainID,
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(rawTx)
}

func SignRawTx(
	inputFile string,
	privateKey *ecdsa.PrivateKey,
	outputFile string,
) error {
	inputFile = ResolveExistingFile(inputFile)
	outputFile = GeneratedFilePath(outputFile)
	if err := EnsureParentDir(outputFile); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	var rawTx map[string]interface{}
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&rawTx); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	// Handle value conversion safely for large numbers
	var value *big.Int
	switch v := rawTx["value"].(type) {
	case float64:
		// For large numbers, we need to handle them carefully
		valueStr := fmt.Sprintf("%.0f", v)
		value = new(big.Int)
		value.SetString(valueStr, 10)
	case string:
		value = new(big.Int)
		value.SetString(v, 10)
	default:
		return fmt.Errorf("invalid value type: %T", v)
	}
	chainID, ok := rawTx["chainId"].(float64)
	if !ok {
		return fmt.Errorf("invalid chainId type: %T", rawTx["chainId"])
	}
	if err := ValidateChainID(int64(chainID)); err != nil {
		PrintValidationError(err)
		return err
	}

	tx := types.NewTransaction(
		uint64(rawTx["nonce"].(float64)),
		common.HexToAddress(rawTx["to"].(string)),
		value,
		uint64(rawTx["gasLimit"].(float64)),
		big.NewInt(int64(rawTx["gasPrice"].(float64))),
		common.Hex2Bytes(rawTx["data"].(string)),
	)

	signer := types.NewEIP155Signer(big.NewInt(int64(chainID)))
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign: %v", err)
	}

	signedData, err := signedTx.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to encode signed tx: %v", err)
	}
	return os.WriteFile(outputFile, signedData, 0644)
}

func CreateAndSendTx(
	caller common.Address,
	contract common.Address,
	value *big.Int,
	data []byte,
	rpc string,
	privateKey *ecdsa.PrivateKey,
) (common.Hash, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to connect to RPC: %v", err)
	}
	defer client.Close()

	tx, chainID, err := buildUnsignedTx(client, caller, contract, value, data)
	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return sendSignedTxWithClient(client, signedTx)
}

func SendSignedTx(rpcURL string, signedTxFile string) (common.Hash, error) {
	signedTxFile = ResolveExistingFile(signedTxFile)

	// Read signed tx
	data, err := os.ReadFile(signedTxFile)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to read file: %v", err)
	}

	var tx types.Transaction
	if err := tx.UnmarshalJSON(data); err != nil {
		return common.Hash{}, fmt.Errorf("invalid signed tx: %v", err)
	}

	return SendSignedTxObject(rpcURL, &tx)
}

func SendSignedTxObject(rpcURL string, tx *types.Transaction) (common.Hash, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to connect to RPC: %v", err)
	}
	defer client.Close()

	return sendSignedTxWithClient(client, tx)
}

func sendSignedTxWithClient(client *ethclient.Client, tx *types.Transaction) (common.Hash, error) {
	tx.ChainId()

	sender, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("invalid signed tx: %v", err)
	}
	// Broadcast
	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Printf("send tx error %v\n", err)
		return common.Hash{}, err
	}

	err, blockHeight := waitEthTxFinished(client, tx.Hash())
	if err != nil {
		return tx.Hash(), err
	}
	time.Sleep(3 * time.Second)
	proposer := sender.Hex()
	fmt.Printf("read sender from signed tx is %s\n", proposer)
	err, _ = QueryProposalId(blockHeight.Uint64(), proposer, client)
	return tx.Hash(), err
}

func waitEthTxFinished(client *ethclient.Client, txhash common.Hash) (error, *big.Int) {
	PrintInfo(fmt.Sprintf("Waiting for transaction confirmation: %s", txhash.String()))
	timeout := time.NewTimer(DefaultTimeout * time.Second)
	ticker := time.NewTicker(DefaultCheckInterval * time.Second)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			return errors.New("transaction confirmation timeout"), nil
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(context.Background(), txhash)
			if err == ethereum.NotFound {
				continue
			} else if err != nil {
				return err, nil
			}
			PrintSuccess(fmt.Sprintf("Transaction confirmed in block %v", receipt.BlockNumber))
			return nil, receipt.BlockNumber
		}
	}
}

// Build proposal ID
// flag true to add candidate, false to remove candidate
func BuildProposalId(from, dest string, flag bool, details string, blockNum uint64, client *ethclient.Client) (string, error) {
	sender := common.HexToAddress(from) // Proposer
	dst := common.HexToAddress(dest)    // Candidate
	instance, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		return "", err
	}
	if blockNum == 0 {
		return "", fmt.Errorf("block number must be positive")
	}

	// proposerNonces is incremented after proposal creation. We use (nonceAtBlock - 1).
	nonceAtBlock, err := instance.ProposerNonces(&bind.CallOpts{BlockNumber: big.NewInt(int64(blockNum))}, sender)
	if err != nil {
		return "", err
	}
	if nonceAtBlock.Sign() == 0 {
		return "", fmt.Errorf("no proposer nonce found at block %d", blockNum)
	}
	nonce := new(big.Int).Sub(nonceAtBlock, big.NewInt(1))
	if nonce.Sign() < 0 {
		return "", fmt.Errorf("invalid proposer nonce at block %d", blockNum)
	}

	return buildIdWithNonce(sender, dst, flag, details, nonce)

}

func buildIdWithNonce(
	sender common.Address,
	dst common.Address,
	flag bool,
	details string,
	nonce *big.Int,
) (string, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return "", err
	}
	boolType, err := abi.NewType("bool", "", nil)
	if err != nil {
		return "", err
	}
	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return "", err
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return "", err
	}

	args := abi.Arguments{
		{Type: addressType},
		{Type: addressType},
		{Type: boolType},
		{Type: stringType},
		{Type: uint256Type},
	}
	encoded, err := args.Pack(sender, dst, flag, details, nonce)
	if err != nil {
		return "", err
	}

	hash := crypto.Keccak256(encoded)
	return hex.EncodeToString(hash), nil
}

// Query contracts proposal ID
func QueryProposalId(blockHeight uint64, proposer string, client *ethclient.Client) (error, string) {
	instance, err := contracts.NewProposal(common.HexToAddress(ProposalContractAddr), client)
	if err != nil {
		fmt.Printf("Failed to instantiate contract: %v", err)
		return err, ""
	}

	// Define query filter
	filterOpts := &bind.FilterOpts{
		Start:   blockHeight,  // Starting block number
		End:     &blockHeight, // End block number (nil means latest block)
		Context: context.Background(),
	}
	// Query event logs
	logs, err := instance.FilterLogCreateProposal(filterOpts, nil, []common.Address{common.HexToAddress(proposer)}, nil)
	if err != nil {
		fmt.Printf("Failed to filter LogCreateProposal events: %v", err)
		return err, ""
	}
	// Iterate through logs
	var proposalId string
	for logs.Next() {
		event := logs.Event
		proposalId = hex.EncodeToString(event.Id[:])
		fmt.Println("--------CreateProposal----------")
		fmt.Printf("Proposal ID: %s\n", proposalId)
		fmt.Printf("Proposer: %s\n", event.Proposer.Hex())
		fmt.Printf("Destination: %s\n", event.Dst.Hex())
		fmt.Printf("Flag: %v\n", event.Flag)
		fmt.Printf("Time: %d\n", event.Time)
		fmt.Printf("Block: %d\n", event.Raw.BlockNumber)
		fmt.Println("-----")
	}
	if logs.Error() != nil {
		fmt.Printf("Error reading logs: %v", logs.Error())
		return logs.Error(), ""
	}

	// Query event logs
	logs1, err := instance.FilterLogCreateConfigProposal(filterOpts, nil, []common.Address{common.HexToAddress(proposer)})
	if err != nil {
		fmt.Printf("Failed to filter LogCreateConfigProposal events: %v", err)
		return err, ""
	}
	// Iterate through logs
	proposalId = "0x"
	for logs1.Next() {
		event := logs1.Event
		proposalId = hex.EncodeToString(event.Id[:])
		fmt.Println("--------CreateConfigProposal----------")
		fmt.Printf("Proposal ID: %s\n", proposalId)
		fmt.Printf("Proposer: %s\n", event.Proposer.Hex())
		fmt.Printf("CID: %s\n", event.Cid)
		fmt.Printf("NewValue: %v\n", event.NewValue)
		fmt.Printf("Time: %d\n", event.Time)
		fmt.Printf("Block: %d\n", event.Raw.BlockNumber)
		fmt.Println("-----")
	}
	if logs.Error() != nil {
		fmt.Printf("Error reading logs: %v", logs.Error())
		return logs.Error(), ""
	}

	return nil, proposalId
}

func ReadKeystoreFile(filepath, password string) (*ecdsa.PrivateKey, error) {
	// 1. Read the keystore file
	keyjson, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore: %v", err)
	}

	// 2. Decrypt with password
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %v (wrong password?)", err)
	}

	return key.PrivateKey, nil
}

func WriteJsonFile(data map[string]interface{}, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func ReadFileToString(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	return string(data), nil
}

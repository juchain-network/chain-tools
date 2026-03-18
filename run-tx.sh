#!/usr/bin/env bash

set -euo pipefail

export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
mkdir -p "$GOCACHE"

# run-tx.sh
# Convenience wrapper for the local generate -> sign -> send flow.
#
# Example:
#   ./run-tx.sh proposal create -p 0xPROPOSER -t 0xTARGET -o add --rpc https://rpc.example -k <hex-private-key>

# Parse command line arguments, separating private key (-k) and --rpc parameter
PRIVATE_KEY=""
RPC_URL=""
OTHER_ARGS=()

# Use Bash array indexing
args=("$@")
for ((i=0; i<${#args[@]}; i++)); do
    arg=${args[$i]}
    case "$arg" in
        -k|--private-key)
            ((i+=1))
            PRIVATE_KEY=${args[$i]:-}
            ;;
        --rpc)
            ((i+=1))
            RPC_URL=${args[$i]:-}
            ;;
        *)
            OTHER_ARGS+=("$arg")
            ;;
    esac
done

# Check if private key is provided
if [[ -z $PRIVATE_KEY ]]; then
    echo "Error: Private key (-k) is required"
    exit 1
fi

# Step 1: Generating raw transaction...
echo "Step 1: Generating raw transaction..."
step1_cmd=(go run . "${OTHER_ARGS[@]}")
if [[ -n $RPC_URL ]]; then
    step1_cmd+=(--rpc "$RPC_URL")
fi
"${step1_cmd[@]}" | tee step1_output.txt

# Extract transaction file name from step 1 output
TX_FILE=$(grep -oP 'Transaction file: \K[^\s]+' step1_output.txt)

if [[ -z $TX_FILE ]]; then
    echo "Error: Failed to extract transaction file name from step 1 output"
    exit 1
fi

echo "Extracted transaction file: $TX_FILE"
echo ""

# Step 2: Signing transaction...
echo "Step 2: Signing transaction..."
SIGNED_TX_FILE="${TX_FILE%.*}_signed.json"
go run . misc sign -f "$TX_FILE" -k "$PRIVATE_KEY" | tee step2_output.txt
echo ""

# Step 3: Sending transaction...
echo "Step 3: Sending transaction..."
step3_cmd=(go run . misc send -f "$SIGNED_TX_FILE")
if [[ -n $RPC_URL ]]; then
    step3_cmd+=(--rpc "$RPC_URL")
fi
echo "Executing: ${step3_cmd[*]}"
"${step3_cmd[@]}" | tee step3_output.txt

# Clean up temporary files
rm -f step1_output.txt step2_output.txt step3_output.txt

echo "All steps completed successfully!"

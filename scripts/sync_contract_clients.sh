#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

die() {
    echo "Error: $*" >&2
    exit 1
}

to_abs_path() {
    local path="$1"
    if [[ "$path" = /* ]]; then
        echo "$path"
    else
        echo "$PWD/${path#./}"
    fi
}

resolve_source_root() {
    local explicit_root="${CONTRACT_CLIENT_SOURCE_ROOT:-}"

    if [[ -n "$explicit_root" ]]; then
        to_abs_path "$explicit_root"
        return 0
    fi

    if [[ -d "$ROOT_DIR/../chain-contract" ]]; then
        to_abs_path "$ROOT_DIR/../chain-contract"
        return 0
    fi

    if [[ -d "$ROOT_DIR/../chain-contracts" ]]; then
        to_abs_path "$ROOT_DIR/../chain-contracts"
        return 0
    fi

    die "contract source root not found; set CONTRACT_CLIENT_SOURCE_ROOT"
}

resolve_abigen() {
    if [[ -n "${ABIGEN:-}" ]]; then
        echo "$ABIGEN"
        return 0
    fi

    local candidate="$1/../chain/build/bin/abigen"
    if [[ -x "$candidate" ]]; then
        echo "$candidate"
        return 0
    fi

    if command -v abigen >/dev/null 2>&1; then
        command -v abigen
        return 0
    fi

    die "abigen not found; set ABIGEN=/path/to/abigen"
}

SOURCE_ROOT="$(resolve_source_root)"
SOURCE_OUT="$(to_abs_path "${CONTRACT_CLIENT_SOURCE_OUT:-$SOURCE_ROOT/out}")"
TARGET_DIR="$(to_abs_path "${CONTRACT_CLIENT_TARGET_DIR:-$ROOT_DIR/contracts}")"
ABIGEN_BIN="$(resolve_abigen "$SOURCE_ROOT")"
CONTRACTS=(Validators Proposal Punish Staking)

command -v jq >/dev/null 2>&1 || die "jq is required"

if [[ "${CONTRACT_CLIENT_BUILD:-0}" == "1" ]]; then
    command -v forge >/dev/null 2>&1 || die "forge is required when CONTRACT_CLIENT_BUILD=1"
    echo "==> Building contracts in $SOURCE_ROOT"
    (
        cd "$SOURCE_ROOT"
        forge build
    )
fi

[[ -d "$SOURCE_ROOT" ]] || die "contract source root not found: $SOURCE_ROOT"
[[ -d "$SOURCE_OUT" ]] || die "contract artifact directory not found: $SOURCE_OUT"
[[ -x "$ABIGEN_BIN" ]] || die "abigen is not executable: $ABIGEN_BIN"

mkdir -p "$TARGET_DIR"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "==> Source root: $SOURCE_ROOT"
echo "==> Artifact dir: $SOURCE_OUT"
echo "==> Target dir: $TARGET_DIR"
echo "==> Abigen: $ABIGEN_BIN"

for contract in "${CONTRACTS[@]}"; do
    artifact="$SOURCE_OUT/${contract}.sol/${contract}.json"
    abi_file="$TMP_DIR/${contract}.abi"
    bin_file="$TMP_DIR/${contract}.bin"
    out_file="$TARGET_DIR/${contract,,}.go"

    [[ -f "$artifact" ]] || die "missing artifact: $artifact"

    jq '.abi' "$artifact" > "$abi_file"
    jq -r '(.bytecode.object // .bytecode // "") | if type == "string" then . else "" end' "$artifact" > "$bin_file"

    [[ -s "$abi_file" ]] || die "empty ABI extracted from: $artifact"
    [[ -s "$bin_file" ]] || die "empty bytecode extracted from: $artifact"

    echo "==> Generating $out_file"
    "$ABIGEN_BIN" \
        --abi="$abi_file" \
        --bin="$bin_file" \
        --pkg=contracts \
        --type="$contract" \
        --out="$out_file"
done

gofmt -w "$TARGET_DIR"/*.go
echo "==> Contract client sync complete"

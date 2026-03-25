# cli Usage Guide

## Overview

`cli` is the standalone JuChain governance and chain-management CLI extracted from `chain-contract`.

It covers five main areas:

- Proposal management
- Validator operations
- Staking operations
- Punish / evidence submission
- Transaction execution through either offline files or direct online send

## Address Model

Current upstream uses three address roles:

- `validator`: cold address for governance and staking ownership
- `signer`: hot address used for block production
- `feeAddr`: reward receiving address

Operational notes:

- `proposal vote` is sent by the `validator` cold address.
- `validator edit --signer` binds or rotates the hot signer.
- Pending signer data does not always mean “not yet effective”. Use the displayed `Current Block`, `Pending Effective Block`, and `Future Effective` fields together.

## Prerequisites

- Go 1.23 or newer
- Access to a JuChain RPC endpoint for online commands
- A private key or keystore file for signing transactions

## Build and Test

Build the binary:

```bash
make build
./build/cli --help
```

Run tests:

```bash
make test
```

Show build metadata:

```bash
./build/cli version
```

## Command Layout

```text
cli
├── proposal
├── validator
├── staking
├── punish
├── misc
└── version
```

The global `--rpc` flag is required for online commands:

```bash
./build/cli proposal list --rpc https://rpc.example
```

## Proposal Management

Create a validator proposal:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example
```

Create a configuration update proposal:

```bash
./build/cli proposal config \
  -p 0xPROPOSER \
  -i 0 \
  -v 86400 \
  --rpc https://rpc.example
```

Vote on a proposal with the validator cold address:

```bash
./build/cli proposal vote \
  -v 0xVALIDATOR \
  -i PROPOSAL_ID \
  -a \
  --rpc https://rpc.example
```

Send a proposal directly online:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

Query proposals:

```bash
./build/cli proposal query -i PROPOSAL_ID --rpc https://rpc.example
./build/cli proposal list --rpc https://rpc.example
./build/cli proposal param -c 0 --rpc https://rpc.example
```

Config IDs:

- `0` `proposalLastingPeriod`
- `1` `punishThreshold`
- `2` `removeThreshold`
- `3` `decreaseRate`
- `4` `withdrawProfitPeriod`
- `5` `blockReward`
- `6` `unbondingPeriod`
- `7` `validatorUnjailPeriod`
- `8` `minValidatorStake`
- `9` `maxValidators`
- `10` `minDelegation`
- `11` `minUndelegation`
- `12` `doubleSignSlashAmount`
- `13` `doubleSignRewardAmount`
- `14` `burnAddress`
- `15` `doubleSignWindow`
- `16` `commissionUpdateCooldown`
- `17` `baseRewardRatio`
- `18` `maxCommissionRate`
- `19` `proposalCooldown`

## Validator Operations

List validators:

```bash
./build/cli validator list --rpc https://rpc.example
```

Query a validator:

```bash
./build/cli validator query \
  -a 0xVALIDATOR \
  --rpc https://rpc.example
```

The validator query output includes:

- effective signer
- pending signer
- pending effective block
- current block
- whether the pending signer is still future-effective

Edit validator fee address only:

```bash
./build/cli validator edit \
  -v 0xVALIDATOR \
  -f 0xFEE_ADDRESS \
  -m my-validator \
  --rpc https://rpc.example
```

Edit validator and rotate signer:

```bash
./build/cli validator edit \
  -v 0xVALIDATOR \
  -f 0xFEE_ADDRESS \
  --signer 0xSIGNER \
  -m my-validator \
  --rpc https://rpc.example
```

Signer-related queries:

```bash
./build/cli validator signer -v 0xVALIDATOR --rpc https://rpc.example
./build/cli validator pending-signer -v 0xVALIDATOR --rpc https://rpc.example
./build/cli validator by-signer --signer 0xSIGNER --rpc https://rpc.example
./build/cli validator signer-history --signer 0xSIGNER --rpc https://rpc.example
./build/cli validator pending-by-signer --signer 0xSIGNER --rpc https://rpc.example
./build/cli validator active-signers --rpc https://rpc.example
./build/cli validator top-signers --rpc https://rpc.example
./build/cli validator epoch-signers --rpc https://rpc.example
./build/cli validator reward-signers --rpc https://rpc.example
```

Signer-set note:

- `top-signers` means the current effective runtime signer set
- `epoch-signers` means the checkpoint/header transition view and is not always identical to the runtime set

Create a validator reward claim transaction:

```bash
./build/cli validator claim \
  -c 0xCALLER \
  -v 0xVALIDATOR \
  --rpc https://rpc.example
```

## Staking Operations

Register a validator:

```bash
./build/cli staking validator-register \
  -p 0xVALIDATOR \
  -s 100000 \
  -c 500 \
  --rpc https://rpc.example
```

Delegate and undelegate:

```bash
./build/cli staking delegate \
  -d 0xDELEGATOR \
  -v 0xVALIDATOR \
  -s 1000 \
  --rpc https://rpc.example

./build/cli staking undelegate \
  -d 0xDELEGATOR \
  -v 0xVALIDATOR \
  -s 500 \
  --rpc https://rpc.example
```

Claim rewards and query delegation state:

```bash
./build/cli staking claim-rewards -c 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking claim-validator-rewards -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking query-delegation -d 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking query-unbonded -d 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
```

Manage validator stake and lifecycle:

```bash
./build/cli staking stake-increase -v 0xVALIDATOR -s 1000 --rpc https://rpc.example
./build/cli staking stake-decrease -v 0xVALIDATOR -s 1000 --rpc https://rpc.example
./build/cli staking set-commission -v 0xVALIDATOR -r 500 --rpc https://rpc.example
./build/cli staking validator-deregister -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking validator-exit -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking validator-unjail -v 0xVALIDATOR --rpc https://rpc.example
./build/cli staking withdraw-unbonded -c 0xCLAIMER -v 0xVALIDATOR --rpc https://rpc.example
```

## Punish Operations

Submit double-sign evidence from hex strings:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1 0xRLP_HEADER_1 \
  --header2 0xRLP_HEADER_2 \
  --rpc https://rpc.example
```

Submit double-sign evidence from files:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1-file /path/to/header1.bin \
  --header2-file /path/to/header2.bin \
  --rpc https://rpc.example
```

Input rules:

- each header must be provided exactly once
- `--headerN` and `--headerN-file` cannot be used together
- file input may be raw bytes or hex text
- the two headers must be conflicting RLP-encoded block headers for the same height

## Offline Transaction Workflow

`cli` follows a three-stage flow for write operations:

1. Generate an unsigned transaction online
2. Sign the JSON payload offline
3. Broadcast the signed transaction later

Example:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example

./build/cli misc sign -f data/createProposal.json -w /path/to/keystore -p /path/to/password.txt

./build/cli misc send -f data/createProposal_signed.json --rpc https://rpc.example
```

You can also sign with a raw hex private key:

```bash
./build/cli misc sign -f data/createProposal.json -k <hex-private-key>
```

Generate a native ETH transfer for the offline flow:

```bash
./build/cli misc transfer \
  -f 0xFROM \
  -t 0xTO \
  -a 0.25 \
  --rpc https://rpc.example
```

Generate a double-sign evidence transaction for later signing:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1-file /path/to/header1.bin \
  --header2-file /path/to/header2.bin \
  --rpc https://rpc.example
```

## Online Send Workflow

For write commands, add `--send` to sign and broadcast in one step without creating raw or signed JSON files.

Using a keystore wallet:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example \
  --send \
  --wallet /path/to/keystore \
  --password /path/to/password.txt
```

Using a raw private key:

```bash
./build/cli staking delegate \
  -d 0xDELEGATOR \
  -v 0xVALIDATOR \
  -s 1000 \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

Native ETH transfer:

```bash
./build/cli misc transfer \
  -f 0xFROM \
  -t 0xTO \
  -a 0.25 \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

Double-sign evidence:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1-file /path/to/header1.bin \
  --header2-file /path/to/header2.bin \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

`--wallet`, `--private-key`, and `--password` only apply when `--send` is enabled.

## Generated Output Files

Common write commands emit JSON files under `data/`, for example:

- `data/createProposal.json`
- `data/createUpdateConfigProposal.json`
- `data/voteProposal.json`
- `data/submitDoubleSignEvidence.json`
- `data/transfer.json`
- `data/registerValidator.json`
- `data/delegate.json`
- `data/claimRewards.json`
- `data/*_signed.json`

These files are created only for the offline flow. The whole `data/` directory is ignored by Git in this repository.

## Contract Binding Sync

The Go files under `contracts/` are generated bindings for the JuChain system contracts.

If contract ABI changes land in `chain-contract`, run:

```bash
make sync-contract-clients
```

## More Help

```bash
./build/cli guide
./build/cli [command] --help
```

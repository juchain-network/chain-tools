# ju-cli Usage Guide

## Overview

`ju-cli` is the standalone JuChain governance and chain-management CLI extracted from `chain-contract`.

It covers four main areas:

- Proposal management
- Validator operations
- Staking operations
- Offline transaction signing and broadcast

## Prerequisites

- Go 1.23 or newer
- Access to a JuChain RPC endpoint for online commands
- A private key or keystore file for signing transactions

## Build and Test

Build the binary:

```bash
make build
./build/ju-cli --help
```

Run tests:

```bash
make test
```

Show build metadata:

```bash
./build/ju-cli version
```

## Command Layout

```text
ju-cli
├── proposal
├── validator
├── staking
├── misc
└── version
```

The global `--rpc` flag is required for online commands:

```bash
./build/ju-cli proposal list --rpc https://rpc.example
```

## Proposal Management

Create a validator proposal:

```bash
./build/ju-cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example
```

Create a configuration update proposal:

```bash
./build/ju-cli proposal config \
  -p 0xPROPOSER \
  -i 0 \
  -v 86400 \
  --rpc https://rpc.example
```

Vote on a proposal:

```bash
./build/ju-cli proposal vote \
  -s 0xVALIDATOR \
  -i PROPOSAL_ID \
  -a \
  --rpc https://rpc.example
```

Query proposals:

```bash
./build/ju-cli proposal query -i PROPOSAL_ID --rpc https://rpc.example
./build/ju-cli proposal list --rpc https://rpc.example
./build/ju-cli proposal param -c 0 --rpc https://rpc.example
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
./build/ju-cli validator list --rpc https://rpc.example
```

Query a validator:

```bash
./build/ju-cli validator query \
  -a 0xVALIDATOR \
  --rpc https://rpc.example
```

Edit validator metadata:

```bash
./build/ju-cli validator edit \
  -v 0xVALIDATOR \
  -f 0xFEE_ADDRESS \
  -m my-validator \
  --rpc https://rpc.example
```

Create a validator reward claim transaction:

```bash
./build/ju-cli validator claim \
  -c 0xCALLER \
  -v 0xVALIDATOR \
  --rpc https://rpc.example
```

## Staking Operations

Register a validator:

```bash
./build/ju-cli staking validator-register \
  -p 0xVALIDATOR \
  -s 100000 \
  -c 500 \
  --rpc https://rpc.example
```

Delegate and undelegate:

```bash
./build/ju-cli staking delegate \
  -d 0xDELEGATOR \
  -v 0xVALIDATOR \
  -s 1000 \
  --rpc https://rpc.example

./build/ju-cli staking undelegate \
  -d 0xDELEGATOR \
  -v 0xVALIDATOR \
  -s 500 \
  --rpc https://rpc.example
```

Claim rewards and query delegation state:

```bash
./build/ju-cli staking claim-rewards -c 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking claim-validator-rewards -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking query-delegation -d 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking query-unbonded -d 0xDELEGATOR -v 0xVALIDATOR --rpc https://rpc.example
```

Manage validator stake and lifecycle:

```bash
./build/ju-cli staking stake-increase -v 0xVALIDATOR -s 1000 --rpc https://rpc.example
./build/ju-cli staking stake-decrease -v 0xVALIDATOR -s 1000 --rpc https://rpc.example
./build/ju-cli staking set-commission -v 0xVALIDATOR -r 500 --rpc https://rpc.example
./build/ju-cli staking validator-deregister -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking validator-exit -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking validator-unjail -v 0xVALIDATOR --rpc https://rpc.example
./build/ju-cli staking withdraw-unbonded -c 0xCLAIMER -v 0xVALIDATOR --rpc https://rpc.example
```

## Offline Transaction Workflow

`ju-cli` follows a three-stage flow for write operations:

1. Generate an unsigned transaction online.
2. Sign the JSON payload offline.
3. Broadcast the signed transaction later.

Example:

```bash
./build/ju-cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example

./build/ju-cli misc sign -f data/createProposal.json -w /path/to/keystore -p /path/to/password.txt

./build/ju-cli misc send -f data/createProposal_signed.json --rpc https://rpc.example
```

You can also sign with a raw hex private key:

```bash
./build/ju-cli misc sign -f data/createProposal.json -k <hex-private-key>
```

## Generated Output Files

Common write commands emit JSON files under `data/`, for example:

- `data/createProposal.json`
- `data/createUpdateConfigProposal.json`
- `data/voteProposal.json`
- `data/registerValidator.json`
- `data/delegate.json`
- `data/claimRewards.json`
- `data/*_signed.json`

The whole `data/` directory is ignored by Git in this repository.

## Contract Binding Sync

The Go files under `contracts/` are generated bindings for the JuChain system contracts.

If contract ABI changes land in `chain-contract`, regenerate the bindings there and sync the updated `contracts/*.go` files into this repository.

## More Help

```bash
./build/ju-cli guide
./build/ju-cli [command] --help
```

# chain-tools

`chain-tools` is the standalone repository for `cli`, the JuChain chain-management and governance command-line tool extracted from `chain-contract`.

## Address Model

Current upstream uses a signer-separated validator model:

- `validator`: cold address for governance and staking ownership
- `signer`: hot address for block production
- `feeAddr`: reward receiving address

Notes:

- `proposal vote` is sent by the validator cold address.
- `validator edit --signer` binds or rotates the hot signer.
- Signer rotation may be recorded on-chain before it becomes runtime-effective; use the validator/signer query commands to inspect `pending signer`, `effective block`, and whether it is still future-effective.

## Scope

- Governance proposals: create, vote, query, and inspect parameter values
- Validator operations: list validators, inspect signer state, edit metadata, and claim validator profits
- Staking operations: register validators, delegate, undelegate, claim rewards, manage stake, and exit flows
- Punish operations: submit double-sign evidence
- Transaction workflow: either use the offline generate/sign/send flow or send directly online with `--send`
- Contract bindings: vendored Go bindings for the JuChain system contracts live under `contracts/`

## Repository Layout

```text
chain-tools/
├── cmd/                  # Cobra commands and shared helpers
├── contracts/            # Generated Go bindings for system contracts
├── docs/cli-guide.md     # Usage guide
├── main.go               # CLI entrypoint
└── Makefile              # Build/test helpers
```

## Quick Start

Build the CLI:

```bash
make build
./build/cli --help
```

Run tests:

```bash
make test
```

Sync generated contract clients from the latest `chain-contract` artifacts:

```bash
make sync-contract-clients
```

Optional overrides:

```bash
make sync-contract-clients CONTRACT_CLIENT_SOURCE_ROOT=../chain-contract
make sync-contract-clients CONTRACT_CLIENT_BUILD=1
make sync-contract-clients ABIGEN=../chain/build/bin/abigen
```

Show version metadata:

```bash
./build/cli version
```

## Common Commands

Create a validator proposal:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
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

Edit validator fee address and rotate signer:

```bash
./build/cli validator edit \
  -v 0xVALIDATOR \
  -f 0xFEE_ADDRESS \
  --signer 0xSIGNER \
  -m my-validator \
  --rpc https://rpc.example
```

Inspect signer state:

```bash
./build/cli validator query -a 0xVALIDATOR --rpc https://rpc.example
./build/cli validator signer -v 0xVALIDATOR --rpc https://rpc.example
./build/cli validator pending-signer -v 0xVALIDATOR --rpc https://rpc.example
./build/cli validator by-signer --signer 0xSIGNER --rpc https://rpc.example
./build/cli validator reward-signers --rpc https://rpc.example
```

Signer-set note:

- `validator top-signers` shows the current effective runtime signer set
- `validator epoch-signers` shows the checkpoint/header transition signer set

Submit double-sign evidence:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1 0xRLP_HEADER_1 \
  --header2 0xRLP_HEADER_2 \
  --rpc https://rpc.example
```

Or read the two headers from files:

```bash
./build/cli punish submit-double-sign \
  -r 0xREPORTER \
  --header1-file /path/to/header1.bin \
  --header2-file /path/to/header2.bin \
  --rpc https://rpc.example
```

Create a native ETH transfer transaction:

```bash
./build/cli misc transfer -f 0xFROM -t 0xTO -a 0.25 --rpc https://rpc.example
```

## Transaction Workflow

Offline flow:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example

./build/cli misc sign -f data/createProposal.json -w /path/to/keystore -p /path/to/password.txt

./build/cli misc send -f data/createProposal_signed.json --rpc https://rpc.example
```

Direct online send:

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

## Notes

- The generated bindings in `contracts/` are tied to the system contract ABI. If ABI changes land in `chain-contract`, regenerate the bindings and sync them here.
- `make` uses a repository-local `GOCACHE` path so builds and tests work in restricted or sandboxed environments too.
- Intermediate transaction JSON files are created under `data/`, and that directory is ignored by Git.
- Double-sign evidence input must be two conflicting RLP-encoded block headers for the same height.

More examples are available in [docs/cli-guide.md](docs/cli-guide.md).

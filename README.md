# chain-tools

`chain-tools` is the standalone repository for `cli`, the JuChain chain management and governance command-line tool extracted from `chain-contract`.

## Scope

- Governance proposals: create, vote, query, and inspect parameter values.
- Validator operations: list validators, query details, edit metadata, and claim validator profits.
- Staking operations: register validators, delegate, undelegate, claim rewards, manage stake, and exit flows.
- Transaction workflow: either use the offline generate/sign/send flow or send directly online with `--send`.
- Contract bindings: vendored Go bindings for the JuChain system contracts live under `contracts/`.

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

Show version metadata:

```bash
./build/cli version
```

## Common Workflow

Create an unsigned proposal transaction:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example
```

Sign it offline:

```bash
./build/cli misc sign -f data/createProposal.json -w /path/to/keystore -p /path/to/password.txt
```

Broadcast the signed transaction:

```bash
./build/cli misc send -f data/createProposal_signed.json --rpc https://rpc.example
```

Create a native ETH transfer transaction:

```bash
./build/cli misc transfer -f 0xFROM -t 0xTO -a 0.25 --rpc https://rpc.example
```

Send directly online with a private key:

```bash
./build/cli proposal create \
  -p 0xPROPOSER \
  -t 0xTARGET_VALIDATOR \
  -o add \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

Send a native ETH transfer directly online:

```bash
./build/cli misc transfer \
  -f 0xFROM \
  -t 0xTO \
  -a 0.25 \
  --rpc https://rpc.example \
  --send \
  --private-key <hex-private-key>
```

More examples are available in [docs/cli-guide.md](docs/cli-guide.md).

## Notes

- The generated bindings in `contracts/` are tied to the system contract ABI. If ABI changes land in `chain-contract`, regenerate the bindings and sync them here.
- `make` uses a repository-local `GOCACHE` path so builds and tests work in restricted or sandboxed environments too.
- Intermediate transaction JSON files are created under `data/`, and that directory is ignored by Git.

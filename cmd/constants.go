package cmd

// Smart contract address constants - JuChain system contract addresses
const (
	// Main contract addresses
	ValidatorContractAddr = "0x000000000000000000000000000000000000f010"
	PunishContractAddr    = "0x000000000000000000000000000000000000f011"
	ProposalContractAddr  = "0x000000000000000000000000000000000000f012"
	StakingContractAddr   = "0x000000000000000000000000000000000000f013"
)

// Default configuration constants
const (
	DefaultGasMultiplier = 120 // Gas estimation multiplier (percentage)
	DefaultTimeout       = 30  // Transaction confirmation timeout (seconds)
	DefaultCheckInterval = 5   // Transaction status check interval (seconds)
	DefaultDataDir       = "data"
)

// Operation type constants
const (
	OperationAdd    = "add"
	OperationRemove = "remove"
)

// Vote type constants
const (
	VoteApprove = true
	VoteReject  = false
)

// Filename template constants - Generic operations
const (
	CreateProposalFile           = "createProposal.json"
	CreateConfigProposalFile     = "createUpdateConfigProposal.json"
	VoteProposalFile             = "voteProposal.json"
	SubmitDoubleSignEvidenceFile = "submitDoubleSignEvidence.json"
	TransferFile                 = "transfer.json"
	WithdrawProfitsFile          = "withdrawProfits.json"
)

// Filename template constants - Staking operations
const (
	RegisterValidatorFile = "registerValidator.json"
	EditValidatorFile     = "editValidator.json"
	DelegateFile          = "delegate.json"
	UndelegateFile        = "undelegate.json"
	ClaimRewardsFile      = "claimRewards.json"
	UnjailValidatorFile   = "unjailValidator.json"
)

// Status code constants
const (
	ValidatorStatusActive   = 1
	ValidatorStatusInactive = 0
	ValidatorStatusJailed   = 2
)

package auth

import "github.com/CoreumFoundation/coreum/x/auth/ante"

// DefaultDeterministicGasRequirements returns default config for deterministic gas
func DefaultDeterministicGasRequirements() ante.DeterministicGasRequirements {
	return ante.DeterministicGasRequirements{
		BankSend:          125000,
		GovSubmitProposal: 150000,
		GovVote:           80000,
	}
}

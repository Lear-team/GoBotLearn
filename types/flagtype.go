package types

import (
	"GoBotPigeon/types/apitypes"
)

// Commands ...
type VerificationFlags struct {
	Verification     bool
	LastCommand      *apitypes.LastUserCommand
	FreshLastCommand bool
}

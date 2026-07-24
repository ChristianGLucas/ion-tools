package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Check whether an Ion document (text or binary, auto-detected) is
// well-formed. Parses every top-level value fully (including descending
// into every container) so a deferred parse error — a malformed decimal, an
// invalid UTF-8 string, an out-of-range timestamp — is caught here rather
// than surfacing later in a different node. valid is false and error
// explains why on any parse failure; value_count reports how many top-level
// values were successfully read before validation concluded (the full count
// when valid).
func ValidateIon(ctx context.Context, ax axiom.Context, input *gen.IonInput) (*gen.ValidationResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.ValidationResult{Valid: false, Error: "data is required"}, nil
	}

	return validateIonData(input.Data), nil
}

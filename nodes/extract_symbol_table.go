package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Extract the local symbol table an Ion document (text or binary,
// auto-detected) carries — the table mapping small integer symbol IDs to
// their text, which is what lets Ion binary encode field names, annotation
// names, and symbol values compactly. Most informative for Ion binary
// input, which always carries one; plain Ion text generally returns an
// empty result unless it explicitly declares a $ion_symbol_table struct —
// that is expected behavior, not an error.
func ExtractSymbolTable(ctx context.Context, ax axiom.Context, input *gen.IonInput) (*gen.SymbolTableInfo, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.SymbolTableInfo{Error: "data is required"}, nil
	}

	info, err := extractSymbolTableFromIon(input.Data)
	if err != nil {
		return &gen.SymbolTableInfo{Error: err.Error()}, nil
	}
	return info, nil
}

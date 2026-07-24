package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Produce a structural, type-aware summary of an Ion document (text or
// binary, auto-detected): a type tree for each top-level value, the
// distinct Ion types and annotations used anywhere in the document, the
// maximum container nesting depth, and the total value count. This
// surfaces what makes Ion inspection different from inspecting JSON — Ion
// values can carry type annotations (e.g. `meters::5`) and a richer type
// system (timestamp, decimal, symbol, blob, clob, sexp) that this node
// reports explicitly rather than collapsing into a generic "object"/"array".
func InspectStructure(ctx context.Context, ax axiom.Context, input *gen.IonInput) (*gen.StructureReport, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.StructureReport{Error: "data is required"}, nil
	}

	report, err := inspectIon(input.Data)
	if err != nil {
		return &gen.StructureReport{Error: err.Error()}, nil
	}
	return report, nil
}

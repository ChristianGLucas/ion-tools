package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Convert an Ion document (text or binary, auto-detected) to JSON text. This
// is intentionally LOSSY: annotations are dropped, symbols become strings,
// decimals and timestamps become JSON strings (not numbers) holding their
// exact Ion text form, and blobs/clobs become base64 strings — see the
// ionutil.go helper's doc comment for the full, honest list of what does not
// survive. If the document has exactly one top-level Ion value, its direct
// JSON translation is returned; otherwise (zero or multiple top-level
// values) the result is a JSON array of them, in document order.
func IonToJson(ctx context.Context, ax axiom.Context, input *gen.IonInput) (*gen.JsonDoc, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.JsonDoc{Error: "data is required"}, nil
	}

	j, err := ionToJSON(input.Data)
	if err != nil {
		return &gen.JsonDoc{Error: err.Error()}, nil
	}
	return &gen.JsonDoc{Json: j}, nil
}

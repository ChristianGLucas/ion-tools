package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Re-serialize an Ion document (text or binary, auto-detected) as indented,
// human-readable Ion text — one value/field per line with nested containers
// indented, unlike the compact single-line form TextToBinary's inverse
// (BinaryToText) produces.
func PrettyPrint(ctx context.Context, ax axiom.Context, input *gen.IonInput) (*gen.IonText, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.IonText{Error: "data is required"}, nil
	}

	text, err := ionPrettyPrint(input.Data)
	if err != nil {
		return &gen.IonText{Error: err.Error()}, nil
	}
	return &gen.IonText{Text: text}, nil
}

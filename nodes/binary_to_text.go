package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Convert an Ion document from its compact binary encoding to its text
// encoding. Every top-level value, container, field name, and annotation in
// the input is preserved exactly — only the byte-level encoding changes.
func BinaryToText(ctx context.Context, ax axiom.Context, input *gen.IonBinary) (*gen.IonText, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.IonText{Error: "data is required"}, nil
	}

	text, err := ionBinaryToText(input.Data)
	if err != nil {
		return &gen.IonText{Error: err.Error()}, nil
	}
	return &gen.IonText{Text: text}, nil
}

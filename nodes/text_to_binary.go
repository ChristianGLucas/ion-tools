package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Convert an Ion document from its text encoding to its compact binary
// encoding (the form starting with the 0xE0 0x01 0x00 0xEA version marker).
// Every top-level value, container, field name, and annotation in the input
// is preserved exactly — only the byte-level encoding changes.
func TextToBinary(ctx context.Context, ax axiom.Context, input *gen.IonText) (*gen.IonBinary, error) {
	if input == nil || input.Text == "" {
		return &gen.IonBinary{Error: "text is required"}, nil
	}

	data, err := ionTextToBinary(input.Text)
	if err != nil {
		return &gen.IonBinary{Error: err.Error()}, nil
	}
	return &gen.IonBinary{Data: data}, nil
}

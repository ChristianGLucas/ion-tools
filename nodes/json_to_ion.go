package nodes

import (
	"context"

	"christiangeorgelucas/ion-tools/axiom"
	gen "christiangeorgelucas/ion-tools/gen"
)

// Convert a JSON document to Ion text. JSON's null/bool/number/string/
// array/object map onto Ion's null/bool/int-or-float/string/list/struct,
// preserving field order. JSON has no annotation, symbol, decimal,
// timestamp, blob, or clob concept, so this never produces one — chain with
// TextToBinary if you need Ion binary output instead of text.
func JsonToIon(ctx context.Context, ax axiom.Context, input *gen.JsonDoc) (*gen.IonText, error) {
	if input == nil || input.Json == "" {
		return &gen.IonText{Error: "json is required"}, nil
	}

	text, err := jsonToIonText(input.Json)
	if err != nil {
		return &gen.IonText{Error: err.Error()}, nil
	}
	return &gen.IonText{Text: text}, nil
}

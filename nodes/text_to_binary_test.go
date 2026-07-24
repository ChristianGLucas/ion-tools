package nodes_test

import (
	"context"
	"encoding/hex"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// This exact Ion text, and its exact binary encoding below, were cross-
// verified byte-for-byte against Python's amazon.ion (amazon-ion/ion-python,
// an entirely independent Ion implementation from amazon-ion/ion-go — same
// spec, different codebase, different maintainers' test suite):
//
//	>>> import amazon.ion.simpleion as si
//	>>> si.dumps(si.loads(text, single_value=True), binary=True).hex()
//
// produces the identical 70-byte sequence asserted below. This is the
// independent-oracle test for TextToBinary: it does not merely check that
// TextToBinary produces *some* binary and round-trips through itself — it
// checks the exact bytes agree with a second, unrelated Ion encoder.
const oracleIonText = `{name:"widget",count:3,tags:["a","b"],price:19.99,active:true,note:null.string}`

const oracleIonBinaryHex = "e00100eaeea48183dea087be9d85636f756e74847461677385707269636586616374697665846e6f7465de9a84867769646765748a21038bb4816181628c53c207cf8d118e8f"

func TestTextToBinary_MatchesIndependentEncoder(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.TextToBinary(context.Background(), ctx, &gen.IonText{Text: oracleIonText})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	gotHex := hex.EncodeToString(got.Data)
	if gotHex != oracleIonBinaryHex {
		t.Fatalf("binary encoding mismatch:\n got:  %s\n want: %s (from independent Python amazon.ion encoder)", gotHex, oracleIonBinaryHex)
	}
}

func TestTextToBinary_MissingTextIsError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.TextToBinary(context.Background(), ctx, &gen.IonText{})
	if err != nil {
		t.Fatalf("unexpected Go error (should be a structured field, not a thrown error): %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for empty text, got none: %+v", got)
	}
}

func TestTextToBinary_MalformedTextIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.TextToBinary(context.Background(), ctx, &gen.IonText{Text: `{a:1,b:[1,2,`}) // unterminated list+struct
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion text, got none: %+v", got)
	}
	if len(got.Data) != 0 {
		t.Fatalf("expected empty Data alongside a set Error, got %d bytes", len(got.Data))
	}
}

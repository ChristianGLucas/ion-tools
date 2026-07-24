package nodes_test

import (
	"context"
	"encoding/hex"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// Decodes the exact same binary fixture text_to_binary_test.go's
// TestTextToBinary_MatchesIndependentEncoder cross-verified against Python's
// amazon.ion. Together the two tests prove TextToBinary and BinaryToText are
// exact inverses of each other AND agree with an independent implementation.
func TestBinaryToText_MatchesIndependentEncoder(t *testing.T) {
	ctx := newTestContext(t)
	data, err := hex.DecodeString(oracleIonBinaryHex)
	if err != nil {
		t.Fatalf("bad test fixture: %v", err)
	}

	got, err := nodes.BinaryToText(context.Background(), ctx, &gen.IonBinary{Data: data})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	want := oracleIonText + "\n"
	if got.Text != want {
		t.Fatalf("text mismatch:\n got:  %q\n want: %q", got.Text, want)
	}
}

func TestBinaryToText_MissingDataIsError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.BinaryToText(context.Background(), ctx, &gen.IonBinary{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for empty data, got none: %+v", got)
	}
}

func TestBinaryToText_MalformedDataIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	// Well-formed Ion binary version marker followed by garbage that is not
	// a valid Ion type descriptor stream.
	got, err := nodes.BinaryToText(context.Background(), ctx, &gen.IonBinary{Data: []byte{0xE0, 0x01, 0x00, 0xEA, 0xFF, 0xFF, 0xFF}})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion binary, got none: %+v", got)
	}
}

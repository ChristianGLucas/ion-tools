package nodes_test

import (
	"context"
	"encoding/hex"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// Decodes the same independently-cross-verified binary fixture used by
// text_to_binary_test.go/binary_to_text_test.go. Its struct has 6 field
// names (name, count, tags, price, active, note) but only 5 become NEW
// local symbols — "name" is already system symbol SID 4 (part of Ion's
// built-in $ion_symbol_table system symbols, reused rather than redeclared),
// so it does not appear in the local symbols list. This is a genuine,
// non-obvious piece of Ion behavior (system symbol table interning), and
// this is a hand-computed known value: the system symbol table's 9 entries
// are documented in the Ion 1.0 spec, so max_id = 9 + 5 new = 14 is a fact
// about the format, not an implementation detail we're guessing at.
func TestExtractSymbolTable_BinaryDocumentLocalSymbols(t *testing.T) {
	ctx := newTestContext(t)
	data, err := hex.DecodeString(oracleIonBinaryHex)
	if err != nil {
		t.Fatalf("bad test fixture: %v", err)
	}

	got, err := nodes.ExtractSymbolTable(context.Background(), ctx, &gen.IonInput{Data: data})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	wantSymbols := []string{"count", "tags", "price", "active", "note"}
	if !equalStrings(got.Symbols, wantSymbols) {
		t.Fatalf("symbols = %v, want %v (\"name\" is system SID 4, not a new local symbol)", got.Symbols, wantSymbols)
	}
	if got.MaxId != 14 {
		t.Fatalf("max_id = %d, want 14 (9 system + 5 local)", got.MaxId)
	}
}

// Plain Ion text with no field names or symbol values still carries Ion's
// built-in 9-entry system symbol table (SIDs 1-9, fixed by the Ion 1.0 spec)
// and nothing more — this exact list is a hand-computed known value, not an
// implementation detail.
func TestExtractSymbolTable_PlainTextHasOnlySystemSymbols(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ExtractSymbolTable(context.Background(), ctx, &gen.IonInput{Data: []byte(`"just a string"`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantSystemSymbols := []string{
		"$ion", "$ion_1_0", "$ion_symbol_table", "name", "version",
		"imports", "symbols", "max_id", "$ion_shared_symbol_table",
	}
	if !equalStrings(got.Symbols, wantSystemSymbols) {
		t.Fatalf("symbols = %v, want the fixed Ion 1.0 system symbol table %v", got.Symbols, wantSystemSymbols)
	}
	if got.MaxId != 9 {
		t.Fatalf("max_id = %d, want 9 (system symbol table only)", got.MaxId)
	}
}

func TestExtractSymbolTable_MalformedInputIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ExtractSymbolTable(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,`)})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion, got none: %+v", got)
	}
}

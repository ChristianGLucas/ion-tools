package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// A hand-computed known-value check of the whole shape: an annotated struct
// containing a scalar, a nested list, and a nested struct. Every field of
// StructureReport is checked against the value that document unambiguously
// implies (there is only one correct type tree for this input).
func TestInspectStructure_TypesAnnotationsAndDepth(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.InspectStructure(context.Background(), ctx, &gen.IonInput{Data: []byte(`meters::{a:1,b:[1,2,{c:"x"}]}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	if len(got.TopLevelValues) != 1 {
		t.Fatalf("expected 1 top-level value, got %d", len(got.TopLevelValues))
	}
	root := got.TopLevelValues[0]
	if root.IonType != "struct" {
		t.Fatalf("root type = %q, want struct", root.IonType)
	}
	if len(root.Annotations) != 1 || root.Annotations[0] != "meters" {
		t.Fatalf("root annotations = %v, want [meters]", root.Annotations)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 fields on root struct, got %d", len(root.Children))
	}

	a, b := root.Children[0], root.Children[1]
	if a.FieldName != "a" || a.IonType != "int" {
		t.Fatalf("field a = %+v, want field_name=a type=int", a)
	}
	if b.FieldName != "b" || b.IonType != "list" {
		t.Fatalf("field b = %+v, want field_name=b type=list", b)
	}
	if len(b.Children) != 3 {
		t.Fatalf("expected 3 list elements in b, got %d", len(b.Children))
	}
	nestedStruct := b.Children[2]
	if nestedStruct.IonType != "struct" || len(nestedStruct.Children) != 1 {
		t.Fatalf("b[2] = %+v, want a 1-field struct", nestedStruct)
	}
	if nestedStruct.Children[0].FieldName != "c" || nestedStruct.Children[0].IonType != "string" {
		t.Fatalf("b[2].c = %+v, want field_name=c type=string", nestedStruct.Children[0])
	}

	wantTypes := []string{"int", "list", "string", "struct"}
	if !equalStrings(got.IonTypesPresent, wantTypes) {
		t.Fatalf("ion_types_present = %v, want %v", got.IonTypesPresent, wantTypes)
	}
	if len(got.AnnotationsUsed) != 1 || got.AnnotationsUsed[0] != "meters" {
		t.Fatalf("annotations_used = %v, want [meters]", got.AnnotationsUsed)
	}
	if got.MaxDepth != 3 {
		t.Fatalf("max_depth = %d, want 3 (struct -> list -> struct)", got.MaxDepth)
	}
	if got.ValueCount != 7 {
		t.Fatalf("value_count = %d, want 7 (root, a, b, b[0], b[1], b[2], b[2].c)", got.ValueCount)
	}
}

// A null container (e.g. null.list) reports IsNull true and has no children,
// distinct from an empty container.
func TestInspectStructure_NullContainerVsEmptyContainer(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.InspectStructure(context.Background(), ctx, &gen.IonInput{Data: []byte(`[null.list, []]`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	root := got.TopLevelValues[0]
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(root.Children))
	}
	if !root.Children[0].IsNull {
		t.Fatalf("element 0 (null.list) should have is_null=true")
	}
	if root.Children[1].IsNull {
		t.Fatalf("element 1 ([]) should have is_null=false")
	}
}

func TestInspectStructure_MalformedInputIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.InspectStructure(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,`)})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion, got none: %+v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

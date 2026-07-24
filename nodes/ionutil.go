// Package nodes: shared Ion <-> Go/JSON conversion helpers used by every node
// in this package. Not a node itself — axiom compiles any non-registered file
// under nodes/ as shared code the handlers call directly.
//
// All conversion goes through github.com/amazon-ion/ion-go's Reader/Writer
// streaming API, which is the library's own general-purpose, type-preserving
// way to walk an arbitrary Ion value stream (see amazon-ion/ion-go's own
// cmd/ion-go "process" subcommand, which uses the same Reader/Writer pattern
// to transcode between Ion encodings).
package nodes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/amazon-ion/ion-go/ion"

	gen "christiangeorgelucas/ion-tools/gen"
)

// ---------------------------------------------------------------------------
// Ion text <-> Ion binary transcoding (byte-level re-encoding only; no data
// is added, removed, or changed — every value, container, field name, and
// annotation from the input reader is copied to the output writer exactly).
// ---------------------------------------------------------------------------

func ionTextToBinary(text string) ([]byte, error) {
	r := ion.NewReaderString(text)
	var buf bytes.Buffer
	w := ion.NewBinaryWriter(&buf)
	if err := transcodeValues(r, w); err != nil {
		return nil, err
	}
	if err := w.Finish(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ionBinaryToText(data []byte) (string, error) {
	r := ion.NewReaderBytes(data)
	var buf bytes.Buffer
	w := ion.NewTextWriter(&buf)
	if err := transcodeValues(r, w); err != nil {
		return "", err
	}
	if err := w.Finish(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ionPrettyPrint re-serializes Ion (text or binary, auto-detected) as
// indented Ion text.
func ionPrettyPrint(data []byte) (string, error) {
	r := ion.NewReaderBytes(data)
	var buf bytes.Buffer
	w := ion.NewTextWriterOpts(&buf, ion.TextWriterPretty)
	if err := transcodeValues(r, w); err != nil {
		return "", err
	}
	if err := w.Finish(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transcodeValues copies every sibling value at the reader's current position
// (and, for containers, all of their descendants) to the writer.
func transcodeValues(r ion.Reader, w ion.Writer) error {
	for r.Next() {
		name, err := r.FieldName()
		if err != nil {
			return err
		}
		if name != nil && name.Text != nil {
			if err := w.FieldName(*name); err != nil {
				return err
			}
		}

		annos, err := r.Annotations()
		if err != nil {
			return err
		}
		if len(annos) > 0 {
			if err := w.Annotations(annos...); err != nil {
				return err
			}
		}

		if err := transcodeOneValue(r, w); err != nil {
			return err
		}
	}
	return r.Err()
}

func transcodeOneValue(r ion.Reader, w ion.Writer) error {
	switch r.Type() {
	case ion.NullType:
		return w.WriteNull()

	case ion.BoolType:
		v, err := r.BoolValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.BoolType)
		}
		return w.WriteBool(*v)

	case ion.IntType:
		size, err := r.IntSize()
		if err != nil {
			return err
		}
		switch size {
		case ion.NullInt:
			return w.WriteNullType(ion.IntType)
		case ion.Int32:
			v, err := r.IntValue()
			if err != nil {
				return err
			}
			return w.WriteInt(int64(*v))
		case ion.Int64:
			v, err := r.Int64Value()
			if err != nil {
				return err
			}
			return w.WriteInt(*v)
		default: // ion.BigInt
			v, err := r.BigIntValue()
			if err != nil {
				return err
			}
			if v == nil {
				return w.WriteNullType(ion.IntType)
			}
			return w.WriteBigInt(v)
		}

	case ion.FloatType:
		v, err := r.FloatValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.FloatType)
		}
		return w.WriteFloat(*v)

	case ion.DecimalType:
		v, err := r.DecimalValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.DecimalType)
		}
		return w.WriteDecimal(v)

	case ion.TimestampType:
		v, err := r.TimestampValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.TimestampType)
		}
		return w.WriteTimestamp(*v)

	case ion.SymbolType:
		v, err := r.SymbolValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.SymbolType)
		}
		return w.WriteSymbol(*v)

	case ion.StringType:
		v, err := r.StringValue()
		if err != nil {
			return err
		}
		if v == nil {
			return w.WriteNullType(ion.StringType)
		}
		return w.WriteString(*v)

	case ion.ClobType:
		v, err := r.ByteValue()
		if err != nil {
			return err
		}
		if v == nil && r.IsNull() {
			return w.WriteNullType(ion.ClobType)
		}
		return w.WriteClob(v)

	case ion.BlobType:
		v, err := r.ByteValue()
		if err != nil {
			return err
		}
		if v == nil && r.IsNull() {
			return w.WriteNullType(ion.BlobType)
		}
		return w.WriteBlob(v)

	case ion.ListType:
		return transcodeContainer(r, w, ion.ListType)
	case ion.SexpType:
		return transcodeContainer(r, w, ion.SexpType)
	case ion.StructType:
		return transcodeContainer(r, w, ion.StructType)

	default:
		return fmt.Errorf("ion: value at this position has an unsupported type (%v)", r.Type())
	}
}

func transcodeContainer(r ion.Reader, w ion.Writer, t ion.Type) error {
	if r.IsNull() {
		return w.WriteNullType(t)
	}
	if err := r.StepIn(); err != nil {
		return err
	}
	switch t {
	case ion.ListType:
		if err := w.BeginList(); err != nil {
			return err
		}
	case ion.SexpType:
		if err := w.BeginSexp(); err != nil {
			return err
		}
	default:
		if err := w.BeginStruct(); err != nil {
			return err
		}
	}
	if err := transcodeValues(r, w); err != nil {
		return err
	}
	if err := r.StepOut(); err != nil {
		return err
	}
	switch t {
	case ion.ListType:
		return w.EndList()
	case ion.SexpType:
		return w.EndSexp()
	default:
		return w.EndStruct()
	}
}

// ---------------------------------------------------------------------------
// Ion -> JSON
//
// LOSSY BY DESIGN — Ion's type system is strictly richer than JSON's. This
// conversion is honest about what it drops or reshapes:
//   - Annotations (e.g. `meters::5`) are DROPPED — JSON has no equivalent.
//   - Symbol values become JSON strings — the symbol/string distinction is lost.
//     A symbol with no resolvable text renders as the placeholder "$<SID>".
//   - Decimal and Timestamp values become JSON STRINGS preserving their exact
//     numeric value and precision/scale (e.g. "1.50", "2026-07-24T00:00:00Z")
//     — NOT JSON numbers. Note this is the exact VALUE, not necessarily the
//     original input's byte-for-byte text: ion-go's own Decimal.String()
//     renormalizes an exponent-form decimal's coefficient/exponent split
//     (e.g. Ion text "1.50d10" round-trips through this as "150d8" — the
//     same value and significant digits, reformatted). Ion decimal exponent
//     syntax ('d') is not valid JSON number syntax, and JSON numbers cannot
//     exactly preserve arbitrary-precision or trailing-zero-significant
//     decimals, so re-typing them as numbers would silently change the
//     value. A consumer that needs a numeric decimal must parse this string
//     itself.
//   - Blob and Clob both become plain base64-encoded JSON strings — the
//     blob/clob/string distinction is lost; a base64 JSON string produced this
//     way is indistinguishable from a native Ion string that happens to look
//     like base64.
//   - Ion float NaN/+Inf/-Inf (JSON has no such numbers) become the JSON
//     strings "nan", "+inf", "-inf" (Ion's own text spelling for them).
//   - A struct with duplicate field names is preserved as-is in the emitted
//     JSON text, but any downstream JSON parser will apply standard
//     last-value-wins duplicate-key handling.
//   - sexp values become JSON arrays, indistinguishable from list.
//
// This is intentionally NOT a round-trip-faithful conversion; JsonToIon does
// not attempt to reverse any of the above.
// ---------------------------------------------------------------------------

// ionToJSON converts an Ion document to JSON text. If the document has
// exactly one top-level value, that value's JSON translation is returned
// directly. If it has zero or more than one top-level value, the result is a
// JSON array of the top-level values' translations, in document order.
func ionToJSON(data []byte) (string, error) {
	r := ion.NewReaderBytes(data)
	var values []string
	for r.Next() {
		var buf bytes.Buffer
		if err := ionValueToJSON(&buf, r); err != nil {
			return "", err
		}
		values = append(values, buf.String())
	}
	if err := r.Err(); err != nil {
		return "", err
	}
	if len(values) == 1 {
		return values[0], nil
	}
	return "[" + strings.Join(values, ",") + "]", nil
}

func ionValueToJSON(buf *bytes.Buffer, r ion.Reader) error {
	if r.IsNull() {
		buf.WriteString("null")
		return nil
	}

	switch r.Type() {
	case ion.NullType:
		buf.WriteString("null")

	case ion.BoolType:
		v, err := r.BoolValue()
		if err != nil {
			return err
		}
		if *v {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case ion.IntType:
		size, err := r.IntSize()
		if err != nil {
			return err
		}
		switch size {
		case ion.Int32:
			v, err := r.IntValue()
			if err != nil {
				return err
			}
			buf.WriteString(strconv.Itoa(*v))
		case ion.Int64:
			v, err := r.Int64Value()
			if err != nil {
				return err
			}
			buf.WriteString(strconv.FormatInt(*v, 10))
		default:
			v, err := r.BigIntValue()
			if err != nil {
				return err
			}
			buf.WriteString(v.String())
		}

	case ion.FloatType:
		v, err := r.FloatValue()
		if err != nil {
			return err
		}
		writeJSONFloat(buf, *v)

	case ion.DecimalType:
		v, err := r.DecimalValue()
		if err != nil {
			return err
		}
		writeJSONString(buf, v.String())

	case ion.TimestampType:
		v, err := r.TimestampValue()
		if err != nil {
			return err
		}
		writeJSONString(buf, v.String())

	case ion.SymbolType:
		v, err := r.SymbolValue()
		if err != nil {
			return err
		}
		writeJSONString(buf, symbolText(v))

	case ion.StringType:
		v, err := r.StringValue()
		if err != nil {
			return err
		}
		writeJSONString(buf, *v)

	case ion.ClobType, ion.BlobType:
		v, err := r.ByteValue()
		if err != nil {
			return err
		}
		writeJSONString(buf, base64.StdEncoding.EncodeToString(v))

	case ion.ListType, ion.SexpType:
		if err := r.StepIn(); err != nil {
			return err
		}
		buf.WriteByte('[')
		first := true
		for r.Next() {
			if !first {
				buf.WriteByte(',')
			}
			first = false
			if err := ionValueToJSON(buf, r); err != nil {
				return err
			}
		}
		if err := r.Err(); err != nil {
			return err
		}
		if err := r.StepOut(); err != nil {
			return err
		}
		buf.WriteByte(']')

	case ion.StructType:
		if err := r.StepIn(); err != nil {
			return err
		}
		buf.WriteByte('{')
		first := true
		for r.Next() {
			if !first {
				buf.WriteByte(',')
			}
			first = false
			name, err := r.FieldName()
			if err != nil {
				return err
			}
			writeJSONString(buf, symbolText(name))
			buf.WriteByte(':')
			if err := ionValueToJSON(buf, r); err != nil {
				return err
			}
		}
		if err := r.Err(); err != nil {
			return err
		}
		if err := r.StepOut(); err != nil {
			return err
		}
		buf.WriteByte('}')

	default:
		return fmt.Errorf("ion: value at this position has an unsupported type (%v)", r.Type())
	}
	return nil
}

// symbolText resolves a SymbolToken to display text: its known text, or the
// "$<SID>" placeholder when the text is unknown (e.g. a shared-symbol-table
// reference this reader has no catalog for).
func symbolText(tok *ion.SymbolToken) string {
	if tok == nil {
		return "$0"
	}
	if tok.Text != nil {
		return *tok.Text
	}
	return fmt.Sprintf("$%d", tok.LocalSID)
}

func writeJSONString(buf *bytes.Buffer, s string) {
	b, _ := json.Marshal(s) // json.Marshal never errors on a string
	buf.Write(b)
}

func writeJSONFloat(buf *bytes.Buffer, f float64) {
	switch {
	case math.IsNaN(f):
		writeJSONString(buf, "nan")
	case math.IsInf(f, 1):
		writeJSONString(buf, "+inf")
	case math.IsInf(f, -1):
		writeJSONString(buf, "-inf")
	default:
		buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
	}
}

// ---------------------------------------------------------------------------
// JSON -> Ion text
//
// The natural asymmetric counterpart of ionToJSON: JSON's four scalar kinds
// map onto Ion's null/bool/int/float/string, and JSON objects/arrays map onto
// Ion struct/list, preserving field order. JSON has no annotation, symbol,
// decimal, timestamp, blob, or clob concept, so this never produces one —
// round-tripping IonToJson's output back through JsonToIon does not
// reconstruct the original Ion document when the original used any of those.
// ---------------------------------------------------------------------------

func jsonToIonText(jsonText string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonText))
	dec.UseNumber()

	var buf bytes.Buffer
	w := ion.NewTextWriter(&buf)

	if err := jsonValueToIon(dec, w); err != nil {
		return "", err
	}

	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return "", fmt.Errorf("json: unexpected trailing data after the first value")
		}
		return "", err
	}

	if err := w.Finish(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func jsonValueToIon(dec *json.Decoder, w ion.Writer) error {
	tok, err := dec.Token()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("json: unexpected end of input")
		}
		return err
	}
	return writeJSONToken(dec, w, tok)
}

func writeJSONToken(dec *json.Decoder, w ion.Writer, tok json.Token) error {
	switch v := tok.(type) {
	case nil:
		return w.WriteNull()
	case bool:
		return w.WriteBool(v)
	case json.Number:
		return writeJSONNumber(w, v)
	case string:
		return w.WriteString(v)
	case json.Delim:
		switch v {
		case '[':
			if err := w.BeginList(); err != nil {
				return err
			}
			for dec.More() {
				if err := jsonValueToIon(dec, w); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil { // consume closing ']'
				return err
			}
			return w.EndList()
		case '{':
			if err := w.BeginStruct(); err != nil {
				return err
			}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return err
				}
				key, ok := keyTok.(string)
				if !ok {
					return fmt.Errorf("json: expected an object key, got %v", keyTok)
				}
				if err := w.FieldName(ion.NewSymbolTokenFromString(key)); err != nil {
					return err
				}
				if err := jsonValueToIon(dec, w); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil { // consume closing '}'
				return err
			}
			return w.EndStruct()
		default:
			return fmt.Errorf("json: unexpected delimiter %q", v)
		}
	default:
		return fmt.Errorf("json: unexpected token type %T", tok)
	}
}

func writeJSONNumber(w ion.Writer, n json.Number) error {
	s := string(n)
	if !strings.ContainsAny(s, ".eE") {
		bi, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return fmt.Errorf("json: %q is not a valid integer", s)
		}
		if bi.IsInt64() {
			return w.WriteInt(bi.Int64())
		}
		return w.WriteBigInt(bi)
	}
	f, err := n.Float64()
	if err != nil {
		return fmt.Errorf("json: %q is not a valid number: %w", s, err)
	}
	return w.WriteFloat(f)
}

// ---------------------------------------------------------------------------
// Structural inspection
// ---------------------------------------------------------------------------

func inspectIon(data []byte) (*gen.StructureReport, error) {
	r := ion.NewReaderBytes(data)
	report := &gen.StructureReport{}
	typesSeen := map[string]bool{}
	annosSeen := map[string]bool{}
	var maxDepth, valueCount int32

	var walk func(depth int32) (*gen.StructureNode, error)
	walk = func(depth int32) (*gen.StructureNode, error) {
		if depth > maxDepth {
			maxDepth = depth
		}
		valueCount++

		node := &gen.StructureNode{
			IonType: ionTypeName(r.Type()),
			IsNull:  r.IsNull(),
		}
		typesSeen[node.IonType] = true

		if r.IsInStruct() {
			name, err := r.FieldName()
			if err != nil {
				return nil, err
			}
			if name != nil {
				node.FieldName = symbolText(name)
			}
		}

		annos, err := r.Annotations()
		if err != nil {
			return nil, err
		}
		for _, a := range annos {
			text := symbolText(&a)
			node.Annotations = append(node.Annotations, text)
			annosSeen[text] = true
		}

		switch r.Type() {
		case ion.ListType, ion.SexpType, ion.StructType:
			if !node.IsNull {
				if err := r.StepIn(); err != nil {
					return nil, err
				}
				for r.Next() {
					child, err := walk(depth + 1)
					if err != nil {
						return nil, err
					}
					node.Children = append(node.Children, child)
				}
				if err := r.Err(); err != nil {
					return nil, err
				}
				if err := r.StepOut(); err != nil {
					return nil, err
				}
			}
		}
		return node, nil
	}

	for r.Next() {
		node, err := walk(0)
		if err != nil {
			return nil, err
		}
		report.TopLevelValues = append(report.TopLevelValues, node)
	}
	if err := r.Err(); err != nil {
		return nil, err
	}

	types := make([]string, 0, len(typesSeen))
	for t := range typesSeen {
		types = append(types, t)
	}
	sort.Strings(types)

	annosList := make([]string, 0, len(annosSeen))
	for a := range annosSeen {
		annosList = append(annosList, a)
	}
	sort.Strings(annosList)

	report.IonTypesPresent = types
	report.AnnotationsUsed = annosList
	report.MaxDepth = maxDepth
	report.ValueCount = valueCount
	return report, nil
}

func ionTypeName(t ion.Type) string {
	switch t {
	case ion.NullType:
		return "null"
	case ion.BoolType:
		return "bool"
	case ion.IntType:
		return "int"
	case ion.FloatType:
		return "float"
	case ion.DecimalType:
		return "decimal"
	case ion.TimestampType:
		return "timestamp"
	case ion.SymbolType:
		return "symbol"
	case ion.StringType:
		return "string"
	case ion.ClobType:
		return "clob"
	case ion.BlobType:
		return "blob"
	case ion.ListType:
		return "list"
	case ion.SexpType:
		return "sexp"
	case ion.StructType:
		return "struct"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Symbol table extraction
// ---------------------------------------------------------------------------

// extractSymbolTableFromIon reads the local symbol table an Ion document
// carries. Ion binary always carries one (built from the system symbol table
// plus any symbols the document itself declares); plain Ion text always
// carries Ion's built-in 9-entry system symbol table and nothing more unless
// it explicitly encodes a $ion_symbol_table struct.
//
// The document is fully validated first (not just up to the first value's
// type descriptor) so that malformed Ion anywhere in the stream is reported
// as a structured error here too, consistent with every other node — a
// symbol table extracted from a document whose later content turns out to
// be garbage would be misleading to report as a plain success.
func extractSymbolTableFromIon(data []byte) (*gen.SymbolTableInfo, error) {
	if vr := validateIonData(data); !vr.Valid {
		return nil, fmt.Errorf("%s", vr.Error)
	}

	r := ion.NewReaderBytes(data)
	r.Next() // a symbol table is only populated once the reader has positioned on a value
	if err := r.Err(); err != nil {
		return nil, err
	}

	st := r.SymbolTable()
	info := &gen.SymbolTableInfo{}
	if st == nil {
		return info, nil
	}

	info.Symbols = st.Symbols()
	info.MaxId = int32(st.MaxID())
	for _, imp := range st.Imports() {
		info.Imports = append(info.Imports, fmt.Sprintf("%s@%d (max_id %d)", imp.Name(), imp.Version(), imp.MaxID()))
	}
	return info, nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

func validateIonData(data []byte) *gen.ValidationResult {
	r := ion.NewReaderBytes(data)
	var count int32
	for r.Next() {
		if err := validateValue(r); err != nil {
			return &gen.ValidationResult{Valid: false, Error: err.Error(), ValueCount: count}
		}
		count++
	}
	if err := r.Err(); err != nil {
		return &gen.ValidationResult{Valid: false, Error: err.Error(), ValueCount: count}
	}
	return &gen.ValidationResult{Valid: true, ValueCount: count}
}

// validateValue forces full materialization of the current value (and, for a
// container, everything inside it) so that any deferred parse error — a
// malformed decimal, an out-of-range timestamp, an invalid UTF-8 string —
// surfaces here instead of only on some later specific accessor call.
func validateValue(r ion.Reader) error {
	switch r.Type() {
	case ion.NullType:
		return nil
	case ion.BoolType:
		_, err := r.BoolValue()
		return err
	case ion.IntType:
		size, err := r.IntSize()
		if err != nil {
			return err
		}
		switch size {
		case ion.Int32:
			_, err = r.IntValue()
		case ion.Int64:
			_, err = r.Int64Value()
		case ion.BigInt:
			_, err = r.BigIntValue()
		}
		return err
	case ion.FloatType:
		_, err := r.FloatValue()
		return err
	case ion.DecimalType:
		_, err := r.DecimalValue()
		return err
	case ion.TimestampType:
		_, err := r.TimestampValue()
		return err
	case ion.SymbolType:
		_, err := r.SymbolValue()
		return err
	case ion.StringType:
		_, err := r.StringValue()
		return err
	case ion.ClobType, ion.BlobType:
		_, err := r.ByteValue()
		return err
	case ion.ListType, ion.SexpType, ion.StructType:
		if r.IsNull() {
			return nil
		}
		if err := r.StepIn(); err != nil {
			return err
		}
		for r.Next() {
			if err := validateValue(r); err != nil {
				return err
			}
		}
		if err := r.Err(); err != nil {
			return err
		}
		return r.StepOut()
	default:
		return fmt.Errorf("ion: value at this position has an unsupported type (%v)", r.Type())
	}
}

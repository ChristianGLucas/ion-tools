# ion-tools

Composable [Axiom](https://axiomide.com) nodes for [Amazon Ion](https://amazon-ion.github.io/ion-docs/) — the richly-typed, self-describing text-and-binary data format used internally by AWS QLDB, DynamoDB, and other AWS services — wrapping [`amazon-ion/ion-go`](https://github.com/amazon-ion/ion-go) (Apache-2.0), the reference Go implementation maintained by the Ion team.

## Use it from your agent or app

Every node in this package is a **live, auto-scaling API endpoint** on the
[Axiom](https://axiomide.com) marketplace — call it from an AI agent or your
own code, with nothing to self-host.

**📦 See it on the marketplace:**
https://dev.axiomide.com/marketplace/christiangeorgelucas/ion-tools@0.1.0

**Hook it up to an AI agent (MCP).** Add Axiom's hosted MCP server to any MCP
client and every node becomes a typed tool your agent can call — search the
catalog, inspect a schema, and invoke it directly.

```bash
# Claude Code
claude mcp add --transport http axiom https://api.axiomide.com/mcp \
  --header "Authorization: Bearer $AXIOM_API_KEY"
```

Claude Desktop, Cursor, or any config-based client:

```json
{
  "mcpServers": {
    "axiom": {
      "type": "http",
      "url": "https://api.axiomide.com/mcp",
      "headers": { "Authorization": "Bearer YOUR_AXIOM_API_KEY" }
    }
  }
}
```

**Call it from the CLI.**

```bash
axiom invoke christiangeorgelucas/ion-tools/TextToBinary --input '{"text":"{name:\"widget\",count:3}"}'
```

**Call it over HTTP.**

```bash
curl -X POST https://api.axiomide.com/invocations/v1/nodes/christiangeorgelucas/ion-tools/0.1.0/TextToBinary \
  -H "Authorization: Bearer $AXIOM_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"text":"{name:\"widget\",count:3}"}'
```

### Get started free

Install the CLI:

```bash
# macOS / Linux — Homebrew
brew install axiomide/tap/axiom

# macOS / Linux — install script
curl -fsSL https://raw.githubusercontent.com/AxiomIDE/axiom-releases/main/install.sh | sh
```

**Windows:** download the `windows/amd64` `.zip` from the
[releases page](https://github.com/AxiomIDE/axiom-releases/releases), unzip
it, and put `axiom.exe` on your `PATH`.

Then `axiom version` to verify, `axiom login` (GitHub or Google) to
authenticate, and create an API key under **Console → API Keys**. Docs and
sign-up at **[axiomide.com](https://axiomide.com)**.

## Nodes

| Node | Input → Output | What it does |
|---|---|---|
| `TextToBinary` | `IonText` → `IonBinary` | Ion text encoding → Ion binary encoding (pure re-encoding, no data change). |
| `BinaryToText` | `IonBinary` → `IonText` | Ion binary encoding → Ion text encoding — exact inverse of `TextToBinary`. |
| `IonToJson` | `IonInput` → `JsonDoc` | Ion (text or binary, auto-detected) → JSON. Deliberately **lossy** — see below. |
| `JsonToIon` | `JsonDoc` → `IonText` | JSON → Ion text. Never produces annotations/symbols/decimals/timestamps/blobs/clobs (JSON has none). |
| `ValidateIon` | `IonInput` → `ValidationResult` | Is this well-formed Ion? Fully parses every value, including nested containers. |
| `InspectStructure` | `IonInput` → `StructureReport` | A type tree per top-level value, plus the distinct Ion types/annotations used, max nesting depth, and value count. |
| `ExtractSymbolTable` | `IonInput` → `SymbolTableInfo` | The document's local symbol table (symbols, imports, max ID) — most informative for Ion binary. |
| `PrettyPrint` | `IonInput` → `IonText` | Indented, human-readable Ion text. |

All input/auto-detect nodes (`IonToJson`, `ValidateIon`, `InspectStructure`, `ExtractSymbolTable`, `PrettyPrint`) accept Ion in **either** text or binary encoding via `IonInput.data` — the encoding is auto-detected from the leading bytes (Ion binary always starts with the 4-byte version marker `0xE0 0x01 0x00 0xEA`).

## Ion ↔ JSON is honestly, deliberately lossy

Ion's type system is strictly richer than JSON's, so `IonToJson`/`JsonToIon` do **not** round-trip faithfully in either direction. `IonToJson`:

- **Drops annotations entirely** — `meters::5` becomes the bare JSON number `5`.
- Turns **symbols into plain JSON strings** — the symbol/string distinction is lost.
- Turns **decimal and timestamp values into JSON strings** holding their exact Ion text (e.g. `"19.990"`, `"2026-07-24T12:00:00Z"`) — **not** JSON numbers, since JSON numbers can't exactly represent arbitrary-precision/trailing-zero-significant decimals and Ion's `d`-exponent syntax isn't valid JSON number syntax.
- Turns **blob and clob into base64-encoded JSON strings**, indistinguishable from a native Ion string that happens to look like base64.
- Turns **sexp into a JSON array**, indistinguishable from list.

`JsonToIon` is the natural asymmetric counterpart: it never produces an annotation, symbol, decimal, timestamp, blob, or clob, because JSON has no such source data. A JSON document made only of the types JSON and Ion share (null/bool/number/string/array/object) round-trips exactly; anything else does not, by design. Every one of these behaviors has a dedicated regression test in the repo — see `nodes/ion_to_json_test.go`.

## Errors

Every node returns a structured `error` field (never a crash) on malformed input — `data`/`text`/`json` is empty whenever `error` is set.

## License

MIT — see [LICENSE](./LICENSE). Wraps [`amazon-ion/ion-go`](https://github.com/amazon-ion/ion-go) (Apache-2.0). `nodes/testdata/ion_tests_strings.ion` is copied from [`amazon-ion/ion-tests`](https://github.com/amazon-ion/ion-tests) (Apache-2.0), the Ion team's own cross-implementation conformance corpus, used as an independent-oracle test fixture.

Built for the [Axiom](https://axiomide.com) marketplace.

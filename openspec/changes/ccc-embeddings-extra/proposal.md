# Proposal: ccc-embeddings-extra

## Why

`ccc index` fails at runtime with "No module named 'sentence_transformers'" even on a fresh install from sentei's own command: cocoindex-code ships local-embedding support behind the `embeddings-local` extra (which pulls `cocoindex[sentence-transformers]`), and sentei installed both packages bare.

## What Changes

- The ccc install command becomes `uv tool install --upgrade "cocoindex-code[embeddings-local]" --prerelease explicit`; the extra pins `cocoindex>=1.0.6,<1.1.0` with sentence-transformers, subsuming the old `--with "cocoindex>=1.0.0a24"` injection.

## Capabilities

### Modified

- `integration-apply-summary`: install completeness for ccc.

## Impact

- `internal/integration/ccc.go`; a contract test pinning the extra.

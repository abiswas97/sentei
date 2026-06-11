# Design: ccc-embeddings-extra

One-line command change. Dependency facts (verified against the installed
uv env): cocoindex-code 0.2.35 declares `embeddings-local` →
`cocoindex[sentence-transformers]<1.1.0,>=1.0.6`; cocoindex 1.0.8 gates
`sentence-transformers>=3.3.1` behind its `sentence-transformers` extra.
Bare installs therefore index-fail on the default local-embedding config.
The `--prerelease explicit` flag stays (harmless, and still correct while
cocoindex ships alphas). The setup/teardown/detect specs are unchanged.

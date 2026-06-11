# Design: ccc-python-pin

One flag. uv honors cwd `.python-version` even for `tool install`; pinning
`--python 3.11` (the integration's declared floor, and the interpreter the
proven-working env was built on) removes the ambient dependency. uv fetches
a managed CPython if none is installed.

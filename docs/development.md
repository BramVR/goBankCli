---
summary: "Development commands, testing expectations, and fixture rules."
read_when:
  - "Running local verification or adding tests."
  - "Changing development workflow docs."
---
# Development

Use the repository Makefile:

```bash
make fmt
make test
make lint
make ci
```

Add tests next to the package they cover. Use testdata for provider fixtures.
Do not commit real bank data, live credentials, or copied bank exports.

Docs are maintained manually. When command behavior changes, update
`docs/commands.md`, the matching topic doc, and README examples when they are
affected. Every doc under `docs/` must keep `summary` and `read_when`
frontmatter so the shared docs inventory can route future work.

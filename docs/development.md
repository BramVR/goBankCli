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
make docs-site
make docs-site-test
make ci
```

Add tests next to the package they cover. Use testdata for provider fixtures.
Do not commit real bank data, live credentials, or copied bank exports.

Docs are maintained manually. When command behavior changes, update
`docs/commands.md`, the matching topic doc, and README examples when they are
affected. Every doc under `docs/` must keep `summary` and `read_when`
frontmatter so the shared docs inventory can route future work.

The Project Site build renders an explicit allowlist of committed docs into
`dist/docs-site`. Keep planning, ADR, agent, and research notes off that
allowlist unless they are intentionally public.

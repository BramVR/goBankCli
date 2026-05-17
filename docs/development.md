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

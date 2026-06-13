---
summary: "gobankcli version flag reference."
read_when:
  - "Checking CLI discoverability behavior."
  - "Updating build or version output."
---
# gobankcli --version

`gobankcli --version` prints the build version and exits before config or
provider setup.

```bash
gobankcli --version
```

The default development build prints `dev`. Release builds can stamp a
different value at link time.

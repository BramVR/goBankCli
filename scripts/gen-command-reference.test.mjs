import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { renderCommand, renderIndex, yamlString } from "./gen-command-reference.mjs";

describe("yamlString", () => {
  it("quotes frontmatter strings safely", () => {
    assert.equal(yamlString('say "hi"\nnow'), '"say \\"hi\\"\\nnow"');
  });
});

describe("renderCommand", () => {
  const sample = {
    name: "query",
    description: "Run a read-only SQL query against the local archive.",
    usage: "gobankcli query <sql> [flags]",
    positional_args: "<sql>",
    flags: [
      { name: "json", type: "bool", default: "false", description: "Emit stable JSON.", required: false },
      { name: "sql", type: "string", default: "", description: "Read-only SELECT | WITH SQL.", required: true },
    ],
  };

  it("renders command frontmatter, usage, arguments, and flags", () => {
    const out = renderCommand(sample, "gobankcli");
    assert.match(out, /^---\nsummary: "gobankcli query command reference\."/);
    assert.ok(out.includes("```bash\ngobankcli query <sql> [flags]\n```"));
    assert.ok(out.includes("- `<sql>`"));
    assert.ok(out.includes("- `--json` (`bool`, default `false`): Emit stable JSON."));
    assert.ok(out.includes("- `--sql` (`string`, required): Read-only SELECT \\| WITH SQL."));
  });
});

describe("renderIndex", () => {
  it("links visible commands and discoverability pages", () => {
    const out = renderIndex(
      [
        { name: "doctor", description: "Check local config.", usage: "gobankcli doctor [flags]" },
        { name: "query", description: "Run SQL.", usage: "gobankcli query <sql> [flags]" },
      ],
      "gobankcli",
    );
    assert.ok(out.includes("- [`gobankcli doctor`](./commands/doctor.md): Check local config."));
    assert.ok(out.includes("- [`gobankcli query`](./commands/query.md): Run SQL."));
    assert.ok(out.includes("[`gobankcli --help`](./commands/help.md)"));
    assert.ok(out.includes("[`gobankcli --version`](./commands/version.md)"));
  });
});

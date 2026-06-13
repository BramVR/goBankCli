#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const EXPECTED_VERSION = 1;
const VALID_COMMAND_NAME = /^[a-z][a-z0-9-]*$/;
const PRESERVED_DOC_FILES = new Set(["help.md", "version.md"]);

export function renderIndex(commands, binary) {
  const lines = [];
  lines.push("---");
  lines.push(`summary: ${yamlString("CLI command reference generated from the binary.")}`);
  lines.push("read_when:");
  lines.push(`  - ${yamlString("Adding or changing commands, flags, or scriptable output.")}`);
  lines.push(`  - ${yamlString("Updating user-facing command docs.")}`);
  lines.push("---");
  lines.push("");
  lines.push("# Commands");
  lines.push("");
  lines.push(`Every user-facing subcommand exposed by \`${binary}\`. Regenerate with \`make docs-commands\`; committed docs must match the CLI surface.`);
  lines.push("");
  lines.push("## Subcommands");
  lines.push("");
  for (const cmd of commands) {
    const description = cmd.description ? `: ${cmd.description}` : "";
    lines.push(`- [\`${binary} ${cmd.name}\`](./commands/${cmd.name}.md)${description}`);
  }
  lines.push("");
  lines.push("## Discoverability");
  lines.push("");
  lines.push(`- [\`${binary} --help\`](./commands/help.md): built-in help and per-command help.`);
  lines.push(`- [\`${binary} --version\`](./commands/version.md): build version output.`);
  lines.push("");
  return lines.join("\n");
}

export function renderCommand(cmd, binary) {
  const lines = [];
  lines.push("---");
  lines.push(`summary: ${yamlString(`${binary} ${cmd.name} command reference.`)}`);
  lines.push("read_when:");
  lines.push(`  - ${yamlString(`Using or changing ${binary} ${cmd.name}.`)}`);
  lines.push(`  - ${yamlString("Updating command flags, usage, or output behavior.")}`);
  lines.push("---");
  lines.push("");
  lines.push(`# ${binary} ${cmd.name}`);
  lines.push("");
  if (cmd.description) lines.push(cmd.description);
  if (cmd.usage) {
    lines.push("");
    lines.push("## Usage");
    lines.push("");
    lines.push("```bash");
    lines.push(cmd.usage);
    lines.push("```");
  }
  if (cmd.positional_args) {
    lines.push("");
    lines.push("## Arguments");
    lines.push("");
    for (const arg of cmd.positional_args.split(/\s+/).filter(Boolean)) {
      lines.push(`- \`${arg}\``);
    }
  }
  if (Array.isArray(cmd.flags) && cmd.flags.length) {
    lines.push("");
    lines.push("## Flags");
    lines.push("");
    for (const flag of cmd.flags) {
      const requirement = flag.required ? "required" : `default ${formatDefault(flag.default)}`;
      const description = flag.description ? `: ${escapeInline(flag.description)}` : "";
      lines.push(`- \`--${flag.name}\` (\`${flag.type}\`, ${requirement})${description}`);
    }
  }
  lines.push("");
  return lines.join("\n");
}

export function yamlString(text) {
  return JSON.stringify(text ?? "");
}

function formatDefault(value) {
  return value === "" || value == null ? "none" : "`" + escapeInline(value) + "`";
}

function escapeInline(text) {
  return String(text ?? "").replace(/\|/g, "\\|").replace(/\r?\n/g, " ");
}

const invokedDirectly = import.meta.url === `file://${process.argv[1]}`;
if (invokedDirectly) {
  await main();
}

async function main() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  if (!chunks.length) {
    console.error("gen-command-reference: no command metadata JSON on stdin");
    process.exit(1);
  }

  let doc;
  try {
    doc = JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch (err) {
    console.error(`gen-command-reference: stdin is not valid JSON (${err.message})`);
    process.exit(1);
  }
  if (doc.version !== EXPECTED_VERSION) {
    console.error(`gen-command-reference: metadata version ${doc.version} not supported (expected ${EXPECTED_VERSION})`);
    process.exit(1);
  }
  if (!doc.binary || !Array.isArray(doc.commands)) {
    console.error("gen-command-reference: metadata must include binary and commands");
    process.exit(1);
  }

  const root = process.cwd();
  const docsDir = path.join(root, "docs");
  const outDir = path.join(docsDir, "commands");
  fs.mkdirSync(outDir, { recursive: true });
  for (const entry of fs.readdirSync(outDir, { withFileTypes: true })) {
    if (entry.isFile() && PRESERVED_DOC_FILES.has(entry.name)) continue;
    fs.rmSync(path.join(outDir, entry.name), { recursive: true, force: true });
  }

  let written = 0;
  for (const cmd of doc.commands) {
    if (!VALID_COMMAND_NAME.test(cmd.name)) {
      console.error(`gen-command-reference: invalid command name ${JSON.stringify(cmd.name)}`);
      process.exit(1);
    }
    const filePath = path.join(outDir, `${cmd.name}.md`);
    fs.writeFileSync(filePath, renderCommand(cmd, doc.binary), "utf8");
    written += 1;
    console.log(`wrote ${path.relative(root, filePath)}`);
  }

  const indexPath = path.join(docsDir, "commands.md");
  fs.writeFileSync(indexPath, renderIndex(doc.commands, doc.binary), "utf8");
  console.log(`wrote ${path.relative(root, indexPath)}`);
  console.log(`generated ${written} command reference pages + index`);
}

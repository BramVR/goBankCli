import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "..");

test("docs-site builds public project-site artifact from allowlisted docs", () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "gobankcli-docs-site-"));
  const outDir = path.join(tempRoot, "out");
  const tempDocs = path.join(tempRoot, "docs");
  fs.cpSync(path.join(root, "docs"), tempDocs, { recursive: true });
  fs.writeFileSync(
    path.join(tempDocs, "research.md"),
    [
      "---",
      'summary: "Internal planning sentinel."',
      "read_when:",
      '  - "Never publish."',
      "---",
      "# Internal Research",
      "",
      "INTERNAL_ONLY_SENTINEL_REAL_BANK_EXPORT",
      "",
    ].join("\n"),
    "utf8",
  );
  fs.appendFileSync(path.join(tempDocs, "index.md"), "\n\n[Unsafe link](javascript:alert(1))\n[Safe link](https://example.com)\n", "utf8");

  try {
    execFileSync("make", ["docs-site"], {
      cwd: root,
      env: {
        ...process.env,
        GOBANKCLI_DOCS_SITE_ROOT: tempRoot,
        GOBANKCLI_DOCS_SITE_OUT: outDir,
        TMPDIR: os.tmpdir(),
      },
      encoding: "utf8",
      stdio: "pipe",
    });
  } finally {
    process.once("exit", () => fs.rmSync(tempRoot, { recursive: true, force: true }));
  }

  const index = fs.readFileSync(path.join(outDir, "index.html"), "utf8");
  const commands = fs.readFileSync(path.join(outDir, "commands.html"), "utf8");
  const dataModel = fs.readFileSync(path.join(outDir, "data-model.html"), "utf8");
  const llms = fs.readFileSync(path.join(outDir, "llms.txt"), "utf8");
  const llmsFull = fs.readFileSync(path.join(outDir, "llms-full.txt"), "utf8");

  assert.match(index, /gobankcli/i);
  assert.match(index, /local-first/i);
  assert.match(index, /https:\/\/gobankcli\.bramvanrompuy\.be\//);
  assert.match(index, /href="commands\.html"/);
  assert.match(index, /id="doc-search"/);
  assert.match(index, /data-theme-toggle/);
  assert.match(index, /id="archive-hero-canvas"/);
  assert.match(index, /three@0\.184\.0\/build\/three\.module\.js/);
  assert.match(index, /sha384-8FCZ1eVO6it4\+pbec2aDtnTrwjWXZLJRC\+MAGCIPDgsYnUrl\/E0A2YlF8ioMKI\/J/);
  assert.match(index, /sha384-dw2ooPewaEIrAgl6oFDBmmBWCE9oW9LxRGcfwZ0hLvEprzo202wXl7vCYHRlSnOT/);
  assert.match(index, /class="feature-pill"/);
  assert.match(index, /matchMedia\("\(max-width:960px\)"\)/);
  assert.match(index, /matchMedia\("\(prefers-reduced-motion: reduce\)"\)/);
  assert.doesNotMatch(commands, /class="toc-l[23]" href="#[^"]+">#/);
  assert.equal(fs.readFileSync(path.join(outDir, "CNAME"), "utf8").trim(), "gobankcli.bramvanrompuy.be");

  for (const rel of ["sitemap.xml", "robots.txt", "llms.txt", "llms-full.txt", "favicon.svg", "social-card.svg", ".nojekyll"]) {
    assert.ok(fs.existsSync(path.join(outDir, rel)), `${rel} should exist`);
  }

  assert.match(llms, /Canonical documentation:/);
  assert.match(llmsFull, /# Commands/);
  assert.doesNotMatch(index + llmsFull, /javascript:alert/);
  assert.match(index, /href="https:\/\/example\.com"/);
  assert.doesNotMatch(dataModel, /<p>IBAN\/name\/currency/);
  assert.doesNotMatch(index + llms + llmsFull, /INTERNAL_ONLY_SENTINEL_REAL_BANK_EXPORT/);
  fs.rmSync(tempRoot, { recursive: true, force: true });
});

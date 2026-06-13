import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

const root = path.resolve(import.meta.dirname, "..");
const outDir = path.join(root, "dist", "docs-site");
const internalDoc = path.join(root, "docs", "research.md");
const homeDoc = path.join(root, "docs", "index.md");

test("docs-site builds public project-site artifact from allowlisted docs", () => {
  fs.rmSync(outDir, { recursive: true, force: true });
  const priorInternalDoc = fs.existsSync(internalDoc) ? fs.readFileSync(internalDoc, "utf8") : null;
  const priorHomeDoc = fs.readFileSync(homeDoc, "utf8");
  fs.writeFileSync(
    internalDoc,
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
  fs.writeFileSync(homeDoc, `${priorHomeDoc}\n\n[Unsafe link](javascript:alert(1))\n`, "utf8");

  try {
    execFileSync("make", ["docs-site"], {
      cwd: root,
      env: { ...process.env, TMPDIR: os.tmpdir() },
      encoding: "utf8",
      stdio: "pipe",
    });
  } finally {
    if (priorInternalDoc === null) fs.rmSync(internalDoc, { force: true });
    else fs.writeFileSync(internalDoc, priorInternalDoc, "utf8");
    fs.writeFileSync(homeDoc, priorHomeDoc, "utf8");
  }

  const index = fs.readFileSync(path.join(outDir, "index.html"), "utf8");
  const dataModel = fs.readFileSync(path.join(outDir, "data-model.html"), "utf8");
  const llms = fs.readFileSync(path.join(outDir, "llms.txt"), "utf8");
  const llmsFull = fs.readFileSync(path.join(outDir, "llms-full.txt"), "utf8");

  assert.match(index, /gobankcli/i);
  assert.match(index, /local-first/i);
  assert.match(index, /https:\/\/gobankcli\.bramvanrompuy\.be\//);
  assert.match(index, /href="commands\.html"/);
  assert.equal(fs.readFileSync(path.join(outDir, "CNAME"), "utf8").trim(), "gobankcli.bramvanrompuy.be");

  for (const rel of ["sitemap.xml", "robots.txt", "llms.txt", "llms-full.txt", "favicon.svg", "social-card.svg", ".nojekyll"]) {
    assert.ok(fs.existsSync(path.join(outDir, rel)), `${rel} should exist`);
  }

  assert.match(llms, /Canonical documentation:/);
  assert.match(llmsFull, /# Commands/);
  assert.doesNotMatch(index, /javascript:alert/);
  assert.doesNotMatch(dataModel, /<p>IBAN\/name\/currency/);
  assert.doesNotMatch(index + llms + llmsFull, /INTERNAL_ONLY_SENTINEL_REAL_BANK_EXPORT/);
});

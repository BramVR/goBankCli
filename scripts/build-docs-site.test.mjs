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
  const providerSetup = fs.readFileSync(path.join(outDir, "provider-setup.html"), "utf8");
  const quickstart = fs.readFileSync(path.join(outDir, "quickstart.html"), "utf8");
  const dataModel = fs.readFileSync(path.join(outDir, "data-model.html"), "utf8");
  const llms = fs.readFileSync(path.join(outDir, "llms.txt"), "utf8");
  const llmsFull = fs.readFileSync(path.join(outDir, "llms-full.txt"), "utf8");

  assert.match(index, /gobankcli/i);
  assert.match(index, /local-first/i);
  assert.match(index, /https:\/\/gobankcli\.bramvanrompuy\.be\//);
  assert.match(index, /href="quickstart\.html"/);
  assert.match(index, /href="install\.html"/);
  assert.match(index, /href="provider-setup\.html"/);
  assert.match(index, /id="doc-search"/);
  assert.match(index, /data-theme-toggle/);
  assert.match(index, /class="hero-art"/);
  assert.match(index, /class="archive-art"/);
  assert.match(index, /Local bank archive flow visual/);
  assert.match(index, /BANK DATA · LOCAL/);
  assert.match(index, /SQLITE DB/);
  assert.match(index, /CSV READY/);
  assert.match(index, /#ef5350/);
  assert.match(index, /#fbbf24/);
  assert.match(index, /#4ade80/);
  assert.doesNotMatch(index, /provider sync -> local archive -> export/);
  assert.match(index, /class="archive-flow"/);
  assert.match(index, /class="archive-dot"/);
  assert.match(index, /@media\(prefers-reduced-motion:reduce\).*archive-flow/s);
  assert.doesNotMatch(index, /archive-hero-canvas|three@|threeHeroModule|new THREE|WebGLRenderer/);
  assert.doesNotMatch(index, /vault-art|Particle vault archive visual|vault-link/);
  assert.match(index, /class="feature-pill"/);
  assert.match(providerSetup, /https:\/\/127\.0\.0\.1:28787\/enablebanking\/callback/);
  assert.match(providerSetup, /Restricted production only returns accounts linked/);
  assert.match(providerSetup, /GOBANKCLI_GOCARDLESS_SECRET_ID/);
  assert.match(providerSetup, /Pending transactions are not archived yet/);
  assert.match(quickstart, /does not scrape/i);
  assert.match(quickstart, /cloud upload/i);
  assert.match(index, /matchMedia\("\(max-width:960px\)"\)/);
  assert.match(index, /@media\(prefers-reduced-motion:reduce\)/);
  assert.doesNotMatch(commands, /class="toc-l[23]" href="#[^"]+">#/);
  assert.equal(fs.readFileSync(path.join(outDir, "CNAME"), "utf8").trim(), "gobankcli.bramvanrompuy.be");

  for (const rel of ["sitemap.xml", "robots.txt", "llms.txt", "llms-full.txt", "favicon.svg", "social-card.svg", ".nojekyll", "install.html", "quickstart.html", "provider-setup.html", "archive-query-export.html"]) {
    assert.ok(fs.existsSync(path.join(outDir, rel)), `${rel} should exist`);
  }

  assert.match(llms, /Canonical documentation:/);
  assert.match(llmsFull, /# Commands/);
  assert.match(llmsFull, /# Provider Setup/);
  assert.doesNotMatch(index + llmsFull, /javascript:alert/);
  assert.match(index, /href="https:\/\/example\.com"/);
  assert.doesNotMatch(dataModel, /<p>IBAN\/name\/currency/);
  assert.doesNotMatch(index + llms + llmsFull, /INTERNAL_ONLY_SENTINEL_REAL_BANK_EXPORT/);
  fs.rmSync(tempRoot, { recursive: true, force: true });
});

test("home hero renders short CTAs and linked capability pills", () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "gobankcli-docs-site-home-"));
  const outDir = path.join(tempRoot, "out");
  fs.cpSync(path.join(root, "docs"), path.join(tempRoot, "docs"), { recursive: true });

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

    const index = fs.readFileSync(path.join(outDir, "index.html"), "utf8");
    const hero = index.match(/<header class="home-hero">([\s\S]*?)<\/header>/)?.[1] || "";
    const actions = hero.match(/<div class="actions">([\s\S]*?)<\/div>/)?.[1] || "";
    const ctaLabels = [...actions.matchAll(/<a\b[^>]*class="btn[^"]*"[^>]*>([^<]+)<\/a>/g)].map((match) => match[1]);
    assert.deepEqual(ctaLabels, ["Quickstart", "Install", "GitHub"]);
    assert.match(actions, /<a class="btn primary" href="quickstart\.html">Quickstart<\/a>/);
    assert.doesNotMatch(actions, /Provider setup|Security model/);

    const featureLinks = [...hero.matchAll(/<a\b[^>]*class="feature-pill"[^>]*href="([^"]+)"[^>]*>([\s\S]*?)<\/a>/g)];
    assert.ok(featureLinks.length >= 6, "all capability pills should be anchors");
    for (const [href] of featureLinks.map((match) => [match[1]])) {
      assert.match(href, /^(quickstart|provider-setup|archive-query-export|commands\/sql|security)\.html(?:#[a-z0-9-]+)?$/);
      assert.ok(fs.existsSync(path.join(outDir, href.split("#")[0])), `${href} should point at a generated docs page`);
    }
    assert.equal((hero.match(/class="feature-pill"/g) || []).length, featureLinks.length);
    assert.doesNotMatch(hero, /<span class="feature-pill"/);
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
});

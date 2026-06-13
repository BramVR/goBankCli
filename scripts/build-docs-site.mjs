#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { brandMarkSvg, css, faviconSvg, js, preThemeScript, shieldSvg, socialCardSvg, themeToggleHtml, threeHeroModule } from "./docs-site-assets.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRoot = path.resolve(scriptDir, "..");
const root = path.resolve(process.env.GOBANKCLI_DOCS_SITE_ROOT || defaultRoot);
const docsDir = path.join(root, "docs");
const outDir = path.resolve(process.env.GOBANKCLI_DOCS_SITE_OUT || path.join(root, "dist", "docs-site"));
const repoBase = "https://github.com/BramVR/goBankCli";
const repoEditBase = `${repoBase}/edit/main/docs`;
const productName = "gobankcli";
const productTagline = "Local-first bank transaction archive.";
const productDescription =
  "A local-first, read-only bank transaction archive CLI for provider API sync, SQLite storage, read-only SQL inspection, and stable CSV exports.";
const installHint = "git clone https://github.com/BramVR/goBankCli.git && cd goBankCli && make build";
const cname = readCname();
const siteBase = cname ? `https://${cname}` : "https://bramvr.github.io/goBankCli";

const sections = [
  ["Start", ["index.md", "configuration.md", "commands.md", "examples.md"]],
  ["Concepts", ["architecture.md", "data-model.md", "providers.md", "security.md", "development.md"]],
];

const allowlist = new Set(sections.flatMap(([, rels]) => rels));
const pages = loadPages().filter((page) => allowlist.has(page.rel));
const pageMap = new Map(pages.map((page) => [page.rel, page]));
const nav = sections
  .map(([name, rels]) => ({ name, pages: rels.map((rel) => pageMap.get(rel)).filter(Boolean) }))
  .filter((section) => section.pages.length);
const orderedPages = nav.flatMap((section) => section.pages);
const sectionByRel = new Map(nav.flatMap((section) => section.pages.map((page) => [page.rel, section.name])));

fs.rmSync(outDir, { recursive: true, force: true });
fs.mkdirSync(outDir, { recursive: true });

for (const page of pages) {
  const html = markdownToHtml(page.markdown, page.rel);
  const toc = tocFromHtml(html);
  const idx = orderedPages.findIndex((candidate) => candidate.rel === page.rel);
  const prev = idx > 0 ? orderedPages[idx - 1] : null;
  const next = idx >= 0 && idx < orderedPages.length - 1 ? orderedPages[idx + 1] : null;
  const target = path.join(outDir, page.outRel);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(target, layout({ page, html, toc, prev, next, sectionName: sectionByRel.get(page.rel) || "Docs" }), "utf8");
}

fs.writeFileSync(path.join(outDir, "favicon.svg"), faviconSvg(), "utf8");
fs.writeFileSync(path.join(outDir, "social-card.svg"), socialCardSvg(), "utf8");
fs.writeFileSync(path.join(outDir, ".nojekyll"), "", "utf8");
if (cname) fs.writeFileSync(path.join(outDir, "CNAME"), `${cname}\n`, "utf8");
fs.writeFileSync(path.join(outDir, "llms.txt"), llmsTxt(), "utf8");
fs.writeFileSync(path.join(outDir, "llms-full.txt"), llmsFullTxt(), "utf8");
fs.writeFileSync(path.join(outDir, "sitemap.xml"), sitemapXml(), "utf8");
fs.writeFileSync(path.join(outDir, "robots.txt"), robotsTxt(), "utf8");
validateLinks(outDir);

console.log(`built docs site: ${path.relative(root, outDir)}`);

function loadPages() {
  return allMarkdown(docsDir).map((file) => {
    const rel = path.relative(docsDir, file).replaceAll(path.sep, "/");
    const raw = fs.readFileSync(file, "utf8");
    const { frontmatter, body } = parseFrontmatter(raw);
    const markdown = sanitizeUnsafeMarkdownLinks(stripDirectives(body));
    return {
      file,
      rel,
      markdown,
      frontmatter,
      title: frontmatter.title || firstHeading(markdown) || titleize(path.basename(rel, ".md")),
      outRel: outPath(rel, frontmatter),
    };
  });
}

function readCname() {
  for (const candidate of [path.join(docsDir, "CNAME"), path.join(root, "CNAME")]) {
    if (!fs.existsSync(candidate)) continue;
    return fs.readFileSync(candidate, "utf8").trim().replace(/^https?:\/\//i, "").replace(/\/.*$/, "");
  }
  return "";
}

function parseFrontmatter(raw) {
  const match = raw.match(/^---\n([\s\S]*?)\n---\n?/);
  if (!match) return { frontmatter: {}, body: raw };
  const frontmatter = {};
  for (const line of match[1].split("\n")) {
    const parsed = line.match(/^([A-Za-z0-9_-]+):\s*(.*?)\s*$/);
    if (!parsed) continue;
    let value = parsed[2].trim();
    if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
      value = value.slice(1, -1);
    }
    frontmatter[parsed[1]] = value;
  }
  return { frontmatter, body: raw.slice(match[0].length) };
}

function stripDirectives(body) {
  return body
    .replace(/\r\n/g, "\n")
    .split("\n")
    .filter((line) => !/^\s*\{:\s*[^}]*\}\s*$/.test(line))
    .map((line) => line.replace(/\s*\{:\s*[^}]*\}\s*$/, ""))
    .join("\n");
}

function sanitizeUnsafeMarkdownLinks(markdown) {
  return markdown.replace(/\]\(\s*([A-Za-z][A-Za-z0-9+.-]*:|\/\/)[^)]*\)/g, (match, scheme) => {
    if (/^(https?:|mailto:|tel:)$/i.test(scheme)) return match;
    return "](#)";
  });
}

function allMarkdown(dir) {
  return fs
    .readdirSync(dir, { withFileTypes: true })
    .flatMap((entry) => {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) return allMarkdown(full);
      return entry.name.endsWith(".md") ? [full] : [];
    })
    .sort();
}

function outPath(rel, frontmatter = {}) {
  if (frontmatter.permalink) {
    const permalink = normalizePermalink(frontmatter.permalink);
    return permalink === "/" ? "index.html" : `${permalink.slice(1)}/index.html`;
  }
  return rel === "index.md" ? "index.html" : rel.replace(/\.md$/, ".html");
}

function normalizePermalink(value) {
  let permalink = String(value || "").trim();
  if (!permalink.startsWith("/")) permalink = `/${permalink}`;
  return permalink.length > 1 && permalink.endsWith("/") ? permalink.slice(0, -1) : permalink;
}

function firstHeading(markdown) {
  return markdown.match(/^#\s+(.+)$/m)?.[1]?.trim();
}

function titleize(input) {
  return input.replaceAll("-", " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

function markdownToHtml(markdown, currentRel) {
  const lines = markdown.replace(/\r\n/g, "\n").split("\n");
  const html = [];
  let paragraph = [];
  let list = null;
  let fence = null;

  const flushParagraph = () => {
    if (!paragraph.length) return;
    html.push(`<p>${inline(paragraph.join(" "), currentRel)}</p>`);
    paragraph = [];
  };
  const closeList = () => {
    if (!list) return;
    html.push(`</${list}>`);
    list = null;
  };
  const splitRow = (line) => {
    let trimmed = line.trim();
    if (trimmed.startsWith("|")) trimmed = trimmed.slice(1);
    if (trimmed.endsWith("|")) trimmed = trimmed.slice(0, -1);
    return trimmed.split("|").map((cell) => cell.trim());
  };

  for (let idx = 0; idx < lines.length; idx++) {
    const line = lines[idx];
    const fenceMatch = line.match(/^```([\w+-]+)?\s*$/);
    if (fenceMatch) {
      flushParagraph();
      closeList();
      if (fence) {
        html.push(`<pre><code class="language-${escapeAttr(fence.lang)}">${escapeHtml(fence.lines.join("\n"))}</code></pre>`);
        fence = null;
      } else {
        fence = { lang: fenceMatch[1] || "text", lines: [] };
      }
      continue;
    }
    if (fence) {
      fence.lines.push(line);
      continue;
    }
    if (list && /^\s{2,}\S/.test(line) && html.at(-1)?.startsWith("<li>")) {
      html[html.length - 1] = html.at(-1).replace(/<\/li>$/, ` ${inline(line.trim(), currentRel)}</li>`);
      continue;
    }
    if (!line.trim()) {
      flushParagraph();
      closeList();
      continue;
    }
    const heading = line.match(/^(#{1,4})\s+(.+)$/);
    if (heading) {
      flushParagraph();
      closeList();
      const level = heading[1].length;
      const text = heading[2].trim();
      const id = slug(text);
      if (level === 1) {
        html.push(`<h1 id="${id}">${inline(text, currentRel)}</h1>`);
      } else {
        html.push(`<h${level} id="${id}"><a class="anchor" href="#${id}" aria-label="Anchor link">#</a>${inline(text, currentRel)}</h${level}>`);
      }
      continue;
    }
    if (line.trimStart().startsWith("|") && /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)+\|?\s*$/.test(lines[idx + 1] || "")) {
      flushParagraph();
      closeList();
      const header = splitRow(line);
      idx += 1;
      const rows = [];
      while (idx + 1 < lines.length && lines[idx + 1].trimStart().startsWith("|")) {
        idx += 1;
        rows.push(splitRow(lines[idx]));
      }
      html.push(`<table><thead><tr>${header.map((cell) => `<th>${inline(cell, currentRel)}</th>`).join("")}</tr></thead><tbody>${rows.map((row) => `<tr>${row.map((cell) => `<td>${inline(cell, currentRel)}</td>`).join("")}</tr>`).join("")}</tbody></table>`);
      continue;
    }
    const bullet = line.match(/^\s*-\s+(.+)$/);
    const numbered = line.match(/^\s*\d+\.\s+(.+)$/);
    if (bullet || numbered) {
      flushParagraph();
      const tag = bullet ? "ul" : "ol";
      if (list && list !== tag) closeList();
      if (!list) {
        list = tag;
        html.push(`<${tag}>`);
      }
      html.push(`<li>${inline((bullet || numbered)[1], currentRel)}</li>`);
      continue;
    }
    paragraph.push(line.trim());
  }
  flushParagraph();
  closeList();
  return html.join("\n");
}

function inline(text, currentRel) {
  const stash = [];
  let out = text.replace(/`([^`]+)`/g, (_, code) => {
    stash.push(`<code>${escapeHtml(code)}</code>`);
    return `\u0000${stash.length - 1}\u0000`;
  });
  out = escapeHtml(out)
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
    .replace(/(^|[^*])\*([^*\s][^*]*?)\*(?!\*)/g, "$1<em>$2</em>")
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, href) => `<a href="${escapeAttr(rewriteHref(href, currentRel))}">${label}</a>`)
    .replace(/&lt;(https?:\/\/[^\s<>]+)&gt;/g, '<a href="$1">$1</a>');
  return out.replace(/\u0000(\d+)\u0000/g, (_, idx) => stash[Number(idx)]);
}

function rewriteHref(href, currentRel) {
  if (/^(https?:|mailto:|tel:|#)/.test(href)) return href;
  if (/^[A-Za-z][A-Za-z0-9+.-]*:/.test(href) || href.startsWith("//")) return "#";
  const [raw, hash = ""] = href.split("#");
  if (!raw) return hash ? `#${hash}` : "";
  if (!raw.endsWith(".md")) return href;
  const targetRel = path.posix.normalize(path.posix.join(path.posix.dirname(currentRel), raw));
  const target = pageMap.get(targetRel);
  if (!target) return href;
  const rewritten = hrefToOutRel(target.outRel, pageMap.get(currentRel)?.outRel || outPath(currentRel));
  return `${rewritten}${hash ? `#${hash}` : ""}`;
}

function tocFromHtml(html) {
  const items = [];
  for (const match of html.matchAll(/<h([23]) id="([^"]+)">([\s\S]*?)<\/h[23]>/g)) {
    const text = match[3]
      .replace(/<a\b(?=[^>]*class="[^"]*\banchor\b[^"]*")[\s\S]*?<\/a>/g, "")
      .replace(/<[^>]+>/g, "");
    items.push({ level: Number(match[1]), id: match[2], text });
  }
  if (items.length < 2) return "";
  return `<aside class="toc"><h2>On this page</h2>${items.map((item) => `<a class="toc-l${item.level}" href="#${item.id}">${escapeHtml(item.text)}</a>`).join("")}</aside>`;
}

function layout({ page, html, toc, prev, next, sectionName }) {
  const home = page.outRel === "index.html";
  const rootPrefix = "../".repeat(page.outRel.split("/").length - 1);
  const title = home ? `${productName} - ${productTagline}` : `${page.title} - ${productName}`;
  const description = page.frontmatter.description || (home ? productDescription : `${page.title} documentation for ${productName}.`);
  const canonicalUrl = pageCanonicalUrl(page);
  const socialImage = `${siteBase}/social-card.svg`;
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${escapeHtml(title)}</title>
  <meta name="description" content="${escapeAttr(description)}">
  <link rel="canonical" href="${escapeAttr(canonicalUrl)}">
  <meta name="robots" content="index, follow">
  <meta name="theme-color" content="#1d4ed8">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="${productName}">
  <meta property="og:title" content="${escapeAttr(title)}">
  <meta property="og:description" content="${escapeAttr(description)}">
  <meta property="og:url" content="${escapeAttr(canonicalUrl)}">
  <meta property="og:image" content="${escapeAttr(socialImage)}">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="${escapeAttr(title)}">
  <meta name="twitter:description" content="${escapeAttr(description)}">
  <meta name="twitter:image" content="${escapeAttr(socialImage)}">
  <script type="application/ld+json">${jsonLd({ page, home, canonicalUrl, description })}</script>
  <link rel="icon" href="${rootPrefix}favicon.svg" type="image/svg+xml">
  <script>${preThemeScript()}</script>
  <style>${css()}</style>
</head>
<body${home ? ' class="home"' : ""}>
  <button class="nav-toggle" type="button" aria-label="Toggle navigation" aria-expanded="false">
    <span aria-hidden="true"></span><span aria-hidden="true"></span><span aria-hidden="true"></span>
  </button>
  ${themeToggleHtml("theme-float")}
  <div class="shell">
    <aside class="sidebar">
      <div class="sidebar-head">
        <a class="brand" href="${hrefToOutRel("index.html", page.outRel)}" aria-label="${productName} docs home">
          <span class="mark">${brandMarkSvg()}</span>
          <span><strong>${productName}</strong><small>bank archive CLI</small></span>
        </a>
        ${themeToggleHtml()}
      </div>
      <label class="search"><span>Search</span><input id="doc-search" type="search" placeholder="sync, export, security"></label>
      <nav>${navHtml(page)}</nav>
      <p class="no-results">No matching pages.</p>
    </aside>
    <main>
      ${home ? homeHero() : standardHero(page, sectionName)}
      <div class="doc-grid">
        <article class="doc">${html}${pagerHtml(prev, next, page.outRel)}</article>
        ${home ? "" : toc}
      </div>
    </main>
  </div>
  <script>${js()}</script>
  ${home ? `<script type="module">${threeHeroModule()}</script>` : ""}
</body>
</html>`;
}

function homeHero() {
  const features = [
    "Provider APIs",
    "SQLite archive",
    "Stable CSV",
    "Read-only SQL",
    "No scraping",
    "No payments",
  ];
  return `<header class="home-hero">
        <div>
          <p class="eyebrow">Local-first · read-only · private by default</p>
          <h1>${productName}</h1>
          <p class="lede">${escapeHtml(productDescription)}</p>
          <div class="actions"><a class="btn primary" href="configuration.html">Configure</a><a class="btn" href="commands.html">Commands</a><a class="btn" href="security.html">Security model</a><a class="btn" href="${repoBase}">GitHub</a></div>
          <div class="feature-row" aria-label="Project capabilities">${features.map((feature) => `<span class="feature-pill">${shieldSvg()}${escapeHtml(feature)}</span>`).join("")}</div>
        </div>
        <div class="hero-stage" aria-label="Interactive local bank archive scene">
          <canvas id="archive-hero-canvas"></canvas>
        </div>
      </header>`;
}

function standardHero(page, sectionName) {
  return `<header class="hero"><p class="eyebrow">${escapeHtml(sectionName)}</p><h1>${escapeHtml(page.title)}</h1><div class="actions"><a class="btn" href="${repoBase}">GitHub</a><a class="btn" href="${repoEditBase}/${page.rel}">Edit page</a></div></header>`;
}

function navHtml(currentPage) {
  return nav
    .map((section) => `<section><h2>${escapeHtml(section.name)}</h2>${section.pages.map((page) => {
      const active = page.rel === currentPage.rel ? " active" : "";
      return `<a class="nav-link${active}" href="${hrefToOutRel(page.outRel, currentPage.outRel)}">${escapeHtml(navTitle(page))}</a>`;
    }).join("")}</section>`)
    .join("");
}

function navTitle(page) {
  return page.rel === "index.md" ? "Overview" : page.title;
}

function pagerHtml(prev, next, currentOutRel) {
  if (!prev && !next) return "";
  const cell = (page, label) =>
    page ? `<a href="${hrefToOutRel(page.outRel, currentOutRel)}"><small>${label}</small>${escapeHtml(page.title)}</a>` : "<span></span>";
  return `<nav class="pager" aria-label="Pager">${cell(prev, "Previous")}${cell(next, "Next")}</nav>`;
}

function pageCanonicalUrl(page) {
  if (page.outRel === "index.html") return `${siteBase}/`;
  const rel = page.outRel.endsWith("/index.html") ? page.outRel.slice(0, -"index.html".length) : page.outRel;
  return `${siteBase}/${rel}`;
}

function hrefToOutRel(targetOutRel, currentOutRel) {
  const currentDir = path.posix.dirname(currentOutRel);
  if (targetOutRel === "index.html") return path.posix.relative(currentDir, ".") || ".";
  return path.posix.relative(currentDir, targetOutRel) || path.posix.basename(targetOutRel);
}

function jsonLd({ page, home, canonicalUrl, description }) {
  const value = home
    ? {
        "@context": "https://schema.org",
        "@type": "SoftwareApplication",
        name: productName,
        description,
        url: canonicalUrl,
        applicationCategory: "DeveloperApplication",
        operatingSystem: "macOS, Linux, Windows",
        codeRepository: repoBase,
      }
    : {
        "@context": "https://schema.org",
        "@type": "TechArticle",
        headline: page.title,
        description,
        url: canonicalUrl,
        isPartOf: { "@type": "WebSite", name: productName, url: `${siteBase}/` },
      };
  return JSON.stringify(value).replace(/</g, "\\u003c");
}

function llmsTxt() {
  return [
    `# ${productName}`,
    "",
    productDescription,
    "",
    "Canonical documentation:",
    ...orderedPages.map((page) => `- ${page.title}: ${pageCanonicalUrl(page)}`),
    "",
    "Install:",
    `- ${installHint}`,
    "",
    `Source: ${repoBase}`,
    "",
    "Guidance for agents:",
    "- Prefer these canonical documentation URLs over README excerpts.",
    "- Fetch only the pages needed for the current task.",
    "",
  ].join("\n");
}

function llmsFullTxt() {
  const blocks = [`# ${productName}`, "", productDescription, ""];
  for (const page of orderedPages) {
    blocks.push("---", `# ${page.title}`, `Source: ${pageCanonicalUrl(page)}`, "", page.markdown.trim(), "");
  }
  return `${blocks.join("\n")}\n`;
}

function sitemapXml() {
  const urls = orderedPages.map((page) => `  <url><loc>${escapeXml(pageCanonicalUrl(page))}</loc></url>`);
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.join("\n")}\n</urlset>\n`;
}

function robotsTxt() {
  return `User-agent: *\nAllow: /\n\nSitemap: ${siteBase}/sitemap.xml\n`;
}

function validateLinks(outputDir) {
  const failures = [];
  for (const file of allHtml(outputDir)) {
    const html = fs.readFileSync(file, "utf8");
    for (const match of html.matchAll(/href="([^"]+)"/g)) {
      const href = match[1];
      if (/^(#|https?:|mailto:|tel:)/.test(href)) continue;
      const [rawPath, anchor = ""] = href.split("#");
      const targetPath = rawPath ? path.resolve(path.dirname(file), rawPath) : file;
      const target = fs.existsSync(targetPath) && fs.statSync(targetPath).isDirectory() ? path.join(targetPath, "index.html") : targetPath;
      if (!fs.existsSync(target)) {
        failures.push(`${path.relative(outputDir, file)}: ${href} missing`);
        continue;
      }
      if (anchor && !fs.readFileSync(target, "utf8").includes(`id="${anchor}"`)) {
        failures.push(`${path.relative(outputDir, file)}: ${href} missing anchor`);
      }
    }
  }
  if (failures.length) throw new Error(`broken docs links:\n${failures.join("\n")}`);
}

function allHtml(dir) {
  return fs
    .readdirSync(dir, { withFileTypes: true })
    .flatMap((entry) => {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) return allHtml(full);
      return entry.name.endsWith(".html") ? [full] : [];
    })
    .sort();
}

function slug(text) {
  return text.toLowerCase().replace(/`/g, "").replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[char]);
}

function escapeAttr(value) {
  return escapeHtml(value);
}

function escapeXml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&apos;" })[char]);
}

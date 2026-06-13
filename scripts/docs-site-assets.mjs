export function css() {
  return `
:root{color-scheme:light;--bg:#f7f9fb;--paper:#fff;--ink:#152033;--text:#334155;--muted:#64748b;--line:#dbe4ee;--soft:#eaf3f1;--accent:#0f766e;--accent-2:#2563eb;--code:#0f172a}
*{box-sizing:border-box}
html{scroll-behavior:smooth;scroll-padding-top:24px}
@media(prefers-reduced-motion:reduce){html{scroll-behavior:auto}}
body{margin:0;background:var(--bg);color:var(--text);font-family:Inter,ui-sans-serif,system-ui,-apple-system,Segoe UI,sans-serif;line-height:1.65}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline;text-underline-offset:.18em}
.shell{display:grid;grid-template-columns:250px minmax(0,1fr);min-height:100vh}
.sidebar{position:sticky;top:0;height:100vh;overflow:auto;padding:26px 20px;background:var(--paper);border-right:1px solid var(--line)}
.brand{display:flex;gap:11px;align-items:center;color:var(--ink);margin-bottom:26px;text-decoration:none}
.mark{width:34px;height:34px;border-radius:8px;display:grid;place-items:center;background:var(--soft);color:var(--accent);border:1px solid #c8ddd8}
.mark svg{width:22px;height:22px}.brand strong{display:block;font-size:1.05rem}.brand small{display:block;color:var(--muted);font-size:.72rem;text-transform:uppercase;letter-spacing:.07em}
nav section{margin:0 0 20px}nav h2{margin:0 0 7px;color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.1em}
.nav-link{display:block;border-radius:6px;padding:5px 9px;color:var(--text);font-size:.9rem}.nav-link:hover,.nav-link.active{background:var(--soft);color:var(--accent);text-decoration:none}
main{max-width:1120px;width:100%;padding:48px clamp(20px,5vw,72px) 76px;margin:0 auto}
.hero{border-bottom:1px solid var(--line);padding-bottom:28px;margin-bottom:28px}.eyebrow{margin:0 0 10px;color:var(--accent);font-size:.75rem;text-transform:uppercase;letter-spacing:.11em;font-weight:700}
h1,h2,h3,h4{color:var(--ink);line-height:1.18}h1{font-size:2.45rem;margin:.1em 0 .35em}h2{font-size:1.55rem;margin:2em 0 .55em}h3{font-size:1.16rem;margin:1.5em 0 .35em}
.lede{font-size:1.12rem;max-width:66ch}.actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:22px}.btn{display:inline-flex;align-items:center;border:1px solid var(--ink);border-radius:8px;padding:10px 15px;font-weight:700;color:var(--ink)}.btn.primary{background:var(--ink);color:#fff}.btn:hover{border-color:var(--accent);color:var(--accent);text-decoration:none}.btn.primary:hover{background:var(--accent);color:#fff}
.doc-grid{display:grid;grid-template-columns:minmax(0,72ch) 210px;gap:46px}.doc{min-width:0;overflow-wrap:break-word}.doc h1:first-child{display:none}
.doc code{font-family:"JetBrains Mono",ui-monospace,SFMono-Regular,Menlo,monospace;background:#edf5f3;border:1px solid #d2e5e0;border-radius:5px;padding:.08em .35em;color:#0f766e}.doc pre{overflow:auto;background:var(--code);color:#e2e8f0;border-radius:8px;padding:14px 17px}.doc pre code{background:transparent;border:0;color:inherit;padding:0}.doc table{border-collapse:collapse;width:100%;font-size:.93rem}.doc th,.doc td{border-bottom:1px solid var(--line);padding:8px;text-align:left;vertical-align:top}.doc th{background:#edf3f8;color:var(--ink)}
.toc{position:sticky;top:28px;align-self:start;border-left:1px solid var(--line);padding-left:14px;font-size:.85rem}.toc h2{font-size:.7rem;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin:0 0 8px}.toc a{display:block;color:var(--muted);padding:3px 0}.toc-l3{padding-left:14px!important}
.pager{display:grid;grid-template-columns:1fr 1fr;gap:12px;border-top:1px solid var(--line);margin-top:42px;padding-top:20px}.pager a{border:1px solid var(--line);border-radius:8px;padding:11px 13px;color:var(--ink)}.pager small{display:block;color:var(--muted);text-transform:uppercase;font-size:.68rem}
@media(max-width:900px){.shell{display:block}.sidebar{position:static;height:auto;border-right:0;border-bottom:1px solid var(--line)}main{padding:30px 18px 52px}.doc-grid{display:block}.toc{display:none}h1{font-size:2rem}}
`;
}

export function brandMarkSvg() {
  return `<svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M4 10.5 12 5l8 5.5v8a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 18.5v-8Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/><path d="M8 13h8M8 16h8M9 10h6" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>`;
}

export function faviconSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="gobankcli"><rect width="64" height="64" rx="14" fill="#152033"/><path d="M14 27 32 15l18 12v18a4 4 0 0 1-4 4H18a4 4 0 0 1-4-4V27Z" fill="#eaf3f1" stroke="#5eead4" stroke-width="3" stroke-linejoin="round"/><path d="M23 32h18M23 39h18M26 25h12" stroke="#0f766e" stroke-width="3" stroke-linecap="round"/></svg>`;
}

export function socialCardSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 630" role="img" aria-label="gobankcli social card"><rect width="1200" height="630" fill="#f7f9fb"/><rect x="70" y="70" width="1060" height="490" rx="26" fill="#ffffff" stroke="#dbe4ee"/><g transform="translate(110 110) scale(3.2)" color="#0f766e">${brandMarkSvg()}</g><text x="110" y="285" font-family="Inter, Arial, sans-serif" font-size="86" font-weight="800" fill="#152033">gobankcli</text><text x="114" y="360" font-family="Inter, Arial, sans-serif" font-size="36" fill="#334155">Local-first, read-only bank transaction archive CLI.</text><text x="114" y="438" font-family="JetBrains Mono, Menlo, monospace" font-size="29" fill="#0f766e">SQLite archive · stable CSV · no scraping · no payments</text></svg>`;
}

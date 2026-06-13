export function css() {
  return `
:root[data-theme="dark"]{
  color-scheme:dark;
  --bg:#0b1016;
  --paper:#111923;
  --paper-2:#172232;
  --ink:#f5f8fb;
  --text:#cbd5e1;
  --muted:#93a4b8;
  --subtle:#617080;
  --line:#233142;
  --line-soft:#182333;
  --soft:rgba(94,161,255,.14);
  --accent:#5ea1ff;
  --accent-strong:#8ab4ff;
  --blue:#5ea1ff;
  --warn:#f5c451;
  --code:#080c12;
  --code-border:#263548;
  --shadow:0 18px 45px rgba(0,0,0,.34);
}
:root,:root[data-theme="light"]{
  color-scheme:light;
  --bg:#f5f7fb;
  --paper:#ffffff;
  --paper-2:#edf2ff;
  --ink:#162032;
  --text:#334155;
  --muted:#627084;
  --subtle:#94a1b2;
  --line:#dce5eb;
  --line-soft:#eef3f6;
  --soft:rgba(29,78,216,.10);
  --accent:#1d4ed8;
  --accent-strong:#1e40af;
  --blue:#1d4ed8;
  --warn:#9a6a13;
  --code:#101827;
  --code-border:#243247;
  --shadow:0 12px 34px rgba(20,32,50,.08);
}
*{box-sizing:border-box}
html{scroll-behavior:smooth;scroll-padding-top:26px}
@media(prefers-reduced-motion:reduce){html{scroll-behavior:auto}*,*::before,*::after{animation:none!important;transition:none!important}}
body{margin:0;background:radial-gradient(circle at top right,var(--paper-2),transparent 34rem),var(--bg);color:var(--text);font-family:Inter,ui-sans-serif,system-ui,-apple-system,Segoe UI,sans-serif;line-height:1.65;-webkit-font-smoothing:antialiased}
a{color:var(--accent);text-decoration:none}
a:hover{text-decoration:underline;text-underline-offset:.22em}
.shell{display:grid;grid-template-columns:268px minmax(0,1fr);min-height:100vh}
.sidebar{position:sticky;top:0;height:100vh;overflow:auto;padding:28px 22px;background:color-mix(in srgb,var(--paper) 94%,transparent);border-right:1px solid var(--line);scrollbar-width:thin;scrollbar-color:var(--line) transparent}
.sidebar-head{display:flex;align-items:center;gap:10px;margin-bottom:22px}
.brand{display:flex;gap:11px;align-items:center;min-width:0;flex:1;color:var(--ink);text-decoration:none}
.brand:hover{text-decoration:none}
.mark{width:34px;height:34px;flex:0 0 34px;border-radius:8px;display:grid;place-items:center;background:linear-gradient(135deg,var(--soft),var(--paper-2));color:var(--accent);border:1px solid color-mix(in srgb,var(--accent) 24%,var(--line));box-shadow:var(--shadow)}
.mark svg{width:22px;height:22px}.brand strong{display:block;font-size:1.06rem;line-height:1.05;color:var(--ink)}.brand small{display:block;color:var(--muted);font-size:.68rem;text-transform:uppercase;letter-spacing:.08em;margin-top:3px}
.theme-toggle{width:34px;height:34px;display:inline-grid;place-items:center;border:1px solid var(--line);border-radius:8px;background:transparent;color:var(--muted);cursor:pointer}
.theme-toggle:hover{border-color:var(--accent);color:var(--accent);background:var(--soft)}
.theme-toggle svg{width:16px;height:16px}.theme-toggle .sun{display:none}:root[data-theme="dark"] .theme-toggle .sun{display:block}:root[data-theme="dark"] .theme-toggle .moon{display:none}
.theme-float{display:none}
.search{display:block;margin:0 0 22px}.search span{display:block;margin-bottom:7px;color:var(--muted);font-size:.67rem;font-weight:700;text-transform:uppercase;letter-spacing:.09em}.search input{width:100%;height:38px;border:1px solid var(--line);border-radius:8px;background:var(--paper);color:var(--text);font:inherit;font-size:.88rem;padding:0 11px;outline:none}.search input:focus{border-color:var(--accent);box-shadow:0 0 0 3px var(--soft)}.search input::placeholder{color:var(--subtle)}
nav section{margin:0 0 19px}nav h2{margin:0 0 7px;color:var(--subtle);font-size:.67rem;text-transform:uppercase;letter-spacing:.11em}
.nav-link{display:block;border-radius:7px;padding:5px 10px;color:var(--text);font-size:.9rem;line-height:1.42}.nav-link:hover{background:var(--line-soft);color:var(--ink);text-decoration:none}.nav-link.active{background:var(--soft);color:var(--accent);font-weight:700}
.no-results{display:none;color:var(--muted);font-size:.86rem;margin-top:-4px}
main{max-width:1180px;width:100%;padding:46px clamp(20px,5vw,72px) 84px;margin:0 auto}
.hero{border-bottom:1px solid var(--line);padding:10px 0 28px;margin-bottom:28px}.eyebrow{margin:0 0 10px;color:var(--accent);font-size:.72rem;text-transform:uppercase;letter-spacing:.11em;font-weight:800}
h1,h2,h3,h4{color:var(--ink);line-height:1.18;letter-spacing:0}h1{font-size:2.42rem;margin:.1em 0 .34em}h2{font-size:1.52rem;margin:2em 0 .55em}h3{font-size:1.12rem;margin:1.55em 0 .35em}h4{font-size:1rem;margin:1.35em 0 .25em}
.lede{font-size:1.12rem;max-width:68ch}.actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:22px}.btn{display:inline-flex;align-items:center;gap:7px;border:1px solid var(--line);border-radius:8px;padding:10px 15px;font-weight:750;color:var(--ink);background:var(--paper)}.btn.primary{background:var(--ink);border-color:var(--ink);color:var(--bg)}.btn:hover{border-color:var(--accent);color:var(--accent);background:var(--soft);text-decoration:none}.btn.primary:hover{background:var(--accent-strong);border-color:var(--accent-strong);color:#fff}
.home-hero{display:grid;grid-template-columns:minmax(0,1.04fr) minmax(280px,.72fr);gap:42px;align-items:center;border-bottom:1px solid var(--line);padding:18px 0 34px;margin-bottom:30px}.home-hero h1{font-size:clamp(2.45rem,5vw,4.2rem);line-height:1.02;margin:0 0 18px;max-width:10ch}.home-hero .lede{font-size:1.17rem}
.hero-stage{position:relative;min-height:360px;overflow:visible;isolation:isolate}
.hero-stage canvas{display:block;width:100%;height:100%;min-height:360px}
.feature-row{display:flex;gap:8px;flex-wrap:wrap;margin-top:18px}.feature-pill{display:inline-flex;align-items:center;gap:7px;border:1px solid var(--line);border-radius:999px;padding:6px 11px;background:var(--paper);color:var(--text);font-size:.82rem;font-weight:650}.feature-pill svg{width:15px;height:15px;color:var(--accent);flex:0 0 15px}
.doc-grid{display:grid;grid-template-columns:minmax(0,72ch) 212px;gap:46px}.doc{min-width:0;overflow-wrap:break-word}.doc h1:first-child{display:none}.doc :is(h2,h3,h4){position:relative}.doc :is(h2,h3,h4) .anchor{position:absolute;left:-1.05em;color:var(--subtle);opacity:0;text-decoration:none}.doc :is(h2,h3,h4):hover .anchor{opacity:.75}
.doc p{margin:0 0 1.08em}.doc ul,.doc ol{padding-left:1.35rem;margin:0 0 1.18em}.doc li{margin:.28em 0}.doc strong{color:var(--ink)}
.doc code{font-family:"JetBrains Mono",ui-monospace,SFMono-Regular,Menlo,monospace;background:var(--line-soft);border:1px solid var(--line);border-radius:5px;padding:.08em .35em;color:var(--accent)}.doc pre{position:relative;overflow:auto;background:var(--code);color:#e2e8f0;border-radius:8px;padding:14px 17px;margin:1.35em 0;border:1px solid var(--code-border)}.doc pre code{display:block;background:transparent;border:0;color:inherit;padding:0;font-size:.88rem;white-space:pre}.copy{position:absolute;top:8px;right:8px;border:1px solid rgba(255,255,255,.18);border-radius:6px;background:rgba(255,255,255,.07);color:#e2e8f0;font:700 .7rem/1 Inter,sans-serif;padding:4px 9px;cursor:pointer;opacity:0}.doc pre:hover .copy,.copy:focus{opacity:1}.copy.copied{background:var(--accent);border-color:var(--accent);opacity:1}
.doc table{border-collapse:collapse;width:100%;font-size:.93rem;margin:1.25em 0}.doc th,.doc td{border-bottom:1px solid var(--line);padding:8px;text-align:left;vertical-align:top}.doc th{background:var(--line-soft);color:var(--ink)}
.toc{position:sticky;top:28px;align-self:start;border-left:1px solid var(--line);padding-left:14px;font-size:.85rem;max-height:calc(100vh - 56px);overflow:auto}.toc h2{font-size:.67rem;text-transform:uppercase;letter-spacing:.1em;color:var(--subtle);margin:0 0 8px}.toc a{display:block;color:var(--muted);padding:3px 0}.toc a:hover{color:var(--accent);text-decoration:none}.toc-l3{padding-left:14px!important}
.pager{display:grid;grid-template-columns:1fr 1fr;gap:12px;border-top:1px solid var(--line);margin-top:42px;padding-top:20px}.pager a{border:1px solid var(--line);border-radius:8px;padding:11px 13px;color:var(--ink);background:var(--paper)}.pager a:hover{border-color:var(--accent);background:var(--soft);text-decoration:none}.pager small{display:block;color:var(--muted);text-transform:uppercase;font-size:.68rem;letter-spacing:.08em}
.nav-toggle{display:none;position:fixed;top:14px;right:14px;z-index:20;width:40px;height:40px;border:1px solid var(--line);border-radius:8px;background:var(--paper);color:var(--ink);box-shadow:var(--shadow);padding:9px;cursor:pointer}.nav-toggle span{display:block;height:2px;background:currentColor;border-radius:2px;margin:5px 0}
@media(max-width:960px){.shell{display:block}.sidebar{position:fixed;inset:0 28% 0 0;max-width:330px;z-index:15;transform:translateX(-102%);transition:transform .2s ease;box-shadow:var(--shadow);pointer-events:none}.sidebar.open{transform:translateX(0);pointer-events:auto}.nav-toggle{display:block}.theme-float{display:inline-grid;position:fixed;top:14px;right:62px;z-index:20;width:40px;height:40px;background:var(--paper);color:var(--ink);box-shadow:var(--shadow)}main{padding:62px 18px 56px}.home-hero{display:block}.hero-stage{margin-top:24px;min-height:310px}.hero-stage canvas{min-height:310px}.doc-grid{display:block}.toc{display:none}h1{font-size:2rem}.home-hero h1{font-size:2.7rem}.doc :is(h2,h3,h4) .anchor{display:none}}
`;
}

export function js() {
  return `
const root=document.documentElement;
function readTheme(){try{return localStorage.getItem("theme")}catch{return null}}
function writeTheme(value){try{localStorage.setItem("theme",value)}catch{}}
function setTheme(value){root.dataset.theme=value;document.querySelectorAll("[data-theme-toggle]").forEach((button)=>button.setAttribute("aria-pressed",value==="dark"?"true":"false"))}
setTheme(root.dataset.theme==="light"?"light":"dark");
document.querySelectorAll("[data-theme-toggle]").forEach((button)=>button.addEventListener("click",()=>{const next=root.dataset.theme==="dark"?"light":"dark";setTheme(next);writeTheme(next)}));
const sidebar=document.querySelector(".sidebar");
const toggle=document.querySelector(".nav-toggle");
const mobileNav=window.matchMedia("(max-width:960px)");
function syncNavA11y(open=sidebar?.classList.contains("open")){if(!sidebar)return;const hidden=mobileNav.matches&&!open;sidebar.toggleAttribute("inert",hidden);sidebar.setAttribute("aria-hidden",hidden?"true":"false")}
function setNav(open){if(!sidebar||!toggle)return;sidebar.classList.toggle("open",open);toggle.setAttribute("aria-expanded",open?"true":"false");syncNavA11y(open)}
toggle?.addEventListener("click",()=>setNav(!sidebar?.classList.contains("open")));
document.addEventListener("keydown",(event)=>{if(event.key==="Escape")setNav(false)});
document.addEventListener("click",(event)=>{if(!sidebar?.classList.contains("open"))return;if(sidebar.contains(event.target)||toggle?.contains(event.target))return;setNav(false)});
document.querySelectorAll(".nav-link").forEach((link)=>link.addEventListener("click",()=>setNav(false)));
mobileNav.addEventListener("change",()=>syncNavA11y());
syncNavA11y(false);
const search=document.getElementById("doc-search");
const empty=document.querySelector(".no-results");
search?.addEventListener("input",()=>{const query=search.value.trim().toLowerCase();let anySection=false;document.querySelectorAll(".sidebar nav section").forEach((section)=>{let anyLink=false;section.querySelectorAll(".nav-link").forEach((link)=>{const match=!query||link.textContent.toLowerCase().includes(query);link.style.display=match?"block":"none";if(match)anyLink=true});section.style.display=anyLink?"block":"none";if(anyLink)anySection=true});if(empty)empty.style.display=anySection?"none":"block"});
document.querySelectorAll(".doc pre").forEach((pre)=>{const button=document.createElement("button");button.type="button";button.className="copy";button.textContent="Copy";button.addEventListener("click",async()=>{try{await navigator.clipboard.writeText(pre.querySelector("code")?.textContent??"");button.textContent="Copied";button.classList.add("copied");setTimeout(()=>{button.textContent="Copy";button.classList.remove("copied")},1300)}catch{button.textContent="Failed";setTimeout(()=>{button.textContent="Copy"},1300)}});pre.appendChild(button)});
`;
}

export function threeHeroModule() {
  return `
async function loadThree() {
  const sources = {
    core: {
      url: "https://cdn.jsdelivr.net/npm/three@0.184.0/build/three.core.js",
      integrity: "sha384-dw2ooPewaEIrAgl6oFDBmmBWCE9oW9LxRGcfwZ0hLvEprzo202wXl7vCYHRlSnOT",
    },
    module: {
      url: "https://cdn.jsdelivr.net/npm/three@0.184.0/build/three.module.js",
      integrity: "sha384-8FCZ1eVO6it4+pbec2aDtnTrwjWXZLJRC+MAGCIPDgsYnUrl/E0A2YlF8ioMKI/J",
    },
  };
  async function sriText(source) {
    const response = await fetch(source.url, { cache: "force-cache", integrity: source.integrity });
    if (!response.ok) throw new Error("Unable to load Three.js");
    return response.text();
  }
  const [coreSource, moduleSource] = await Promise.all([sriText(sources.core), sriText(sources.module)]);
  const coreUrl = URL.createObjectURL(new Blob([coreSource], { type: "text/javascript" }));
  const moduleUrl = URL.createObjectURL(
    new Blob([moduleSource.replaceAll("from './three.core.js'", "from " + JSON.stringify(coreUrl))], { type: "text/javascript" }),
  );
  return import(moduleUrl);
}

const THREE = await loadThree();

const canvas = document.getElementById("archive-hero-canvas");
const stage = canvas?.closest(".hero-stage");
if (canvas && stage) {
  const styles = getComputedStyle(document.documentElement);
  const color = (name, fallback) => new THREE.Color(styles.getPropertyValue(name).trim() || fallback);
  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(34, 1, 0.1, 100);
  camera.position.set(4.2, 3.2, 6.2);
  camera.lookAt(0, 0, 0);
  const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
  const root = new THREE.Group();
  scene.add(root);

  const accent = color("--accent", "#1d4ed8");
  const accentStrong = color("--accent-strong", "#1e40af");
  const ink = color("--ink", "#162032");
  const line = color("--line", "#dce5eb");
  const warn = color("--warn", "#f5c451");
  const paper = color("--paper", "#ffffff");

  scene.add(new THREE.HemisphereLight(paper, ink, 2.2));
  const key = new THREE.DirectionalLight(0xffffff, 2.4);
  key.position.set(3, 5, 4);
  scene.add(key);

  const coreMat = new THREE.MeshBasicMaterial({ color: accentStrong, transparent: true, opacity: 0.12, wireframe: true, blending: THREE.AdditiveBlending });
  const diskMat = new THREE.MeshBasicMaterial({ color: accentStrong, transparent: true, opacity: 0.34, wireframe: true, blending: THREE.AdditiveBlending });
  const warnMat = new THREE.MeshBasicMaterial({ color: warn, transparent: true, opacity: 0.9, blending: THREE.AdditiveBlending });
  const accentLine = new THREE.LineBasicMaterial({ color: accentStrong, transparent: true, opacity: 0.56, blending: THREE.AdditiveBlending });

  const shell = new THREE.Mesh(new THREE.CylinderGeometry(0.56, 0.56, 1.5, 64, 1, true), coreMat);
  root.add(shell);

  const disks = [];
  for (let index = 0; index < 4; index += 1) {
    const disk = new THREE.Mesh(new THREE.CylinderGeometry(0.5, 0.5, 0.035, 64), diskMat.clone());
    disk.position.y = -0.38 + index * 0.25;
    disks.push(disk);
    root.add(disk);
  }

  const verticals = [];
  for (let index = 0; index < 8; index += 1) {
    const angle = (index / 8) * Math.PI * 2;
    const x = Math.cos(angle) * 0.56;
    const z = Math.sin(angle) * 0.56;
    const lineGeometry = new THREE.BufferGeometry().setFromPoints([
      new THREE.Vector3(x, -0.78, z),
      new THREE.Vector3(x, 0.78, z),
    ]);
    const lineMesh = new THREE.Line(lineGeometry, accentLine.clone());
    verticals.push(lineMesh);
    root.add(lineMesh);
  }

  const particles = [];
  const particleGeo = new THREE.SphereGeometry(0.035, 16, 8);
  for (let index = 0; index < 10; index += 1) {
    const dot = new THREE.Mesh(particleGeo, index % 5 === 0 ? warnMat : diskMat.clone());
    dot.userData.phase = index / 10;
    dot.userData.radius = 0.76 + (index % 4) * 0.19;
    dot.userData.lift = -0.34 + (index % 5) * 0.17;
    particles.push(dot);
    root.add(dot);
  }

  let pointerX = 0.12;
  let pointerY = 0;
  let pulse = 0;
  let frame = 0;
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
  stage.addEventListener("pointermove", (event) => {
    const rect = stage.getBoundingClientRect();
    pointerX = ((event.clientX - rect.left) / rect.width - 0.5) * 1.15;
    pointerY = ((event.clientY - rect.top) / rect.height - 0.5) * 0.72;
    if (reduceMotion.matches) renderFrame(0, false);
  });
  stage.addEventListener("pointerdown", () => {
    pulse = 1;
    if (reduceMotion.matches) renderFrame(0, false);
  });

  function resize() {
    const rect = stage.getBoundingClientRect();
    renderer.setSize(Math.max(1, rect.width), Math.max(1, rect.height), false);
    camera.aspect = Math.max(1, rect.width) / Math.max(1, rect.height);
    camera.updateProjectionMatrix();
  }

  const observer = new ResizeObserver(resize);
  observer.observe(stage);
  resize();
  stage.dataset.ready = "true";

  function renderFrame(time, moving = true) {
    const t = moving ? time * 0.001 : 0;
    pulse *= moving ? 0.93 : 0;
    root.rotation.y = (moving ? Math.sin(t * 0.42) * 0.28 : 0.08) + pointerX;
    root.rotation.x = -0.08 + pointerY;
    shell.rotation.y = t * 0.18 + pulse * 0.18;
    disks.forEach((disk, index) => {
      disk.rotation.y = t * (0.18 + index * 0.015);
      disk.scale.setScalar(1 + Math.sin(t * 2.1 + index) * 0.025 + pulse * 0.08);
      disk.material.opacity = 0.24 + Math.sin(t * 2 + index) * 0.05 + pulse * 0.16;
    });
    verticals.forEach((lineMesh, index) => {
      lineMesh.material.opacity = 0.28 + Math.sin(t * 2.4 + index * 0.8) * 0.14 + pulse * 0.12;
    });
    particles.forEach((dot, index) => {
      const phase = t * (0.34 + (index % 3) * 0.035) + dot.userData.phase * Math.PI * 2;
      const radius = dot.userData.radius + Math.sin(t + index) * 0.035 + pulse * 0.12;
      dot.position.set(Math.cos(phase) * radius, dot.userData.lift + Math.sin(phase * 1.6) * 0.16, Math.sin(phase) * radius);
      dot.scale.setScalar(0.78 + Math.sin(t * 3 + index) * 0.16 + pulse * 1.1);
      dot.material.opacity = index % 5 === 0 ? 0.82 : 0.34 + pulse * 0.22;
    });
    renderer.render(scene, camera);
  }

  function animate(time) {
    renderFrame(time, true);
    if (!reduceMotion.matches) frame = requestAnimationFrame(animate);
  }

  function startAnimation() {
    cancelAnimationFrame(frame);
    if (reduceMotion.matches) {
      renderFrame(0, false);
    } else {
      frame = requestAnimationFrame(animate);
    }
  }

  reduceMotion.addEventListener("change", startAnimation);
  startAnimation();
}
`;
}

export function preThemeScript() {
  return `(function(){var t;try{t=localStorage.getItem("theme")}catch(e){}document.documentElement.dataset.theme=t==="light"?"light":"dark"})();`;
}

export function themeToggleHtml(extraClass = "") {
  const className = extraClass ? `theme-toggle ${extraClass}` : "theme-toggle";
  return `<button class="${className}" type="button" aria-label="Toggle dark mode" aria-pressed="true" data-theme-toggle>
    <svg class="moon" viewBox="0 0 20 20" aria-hidden="true"><path d="M14.6 12.1A6.5 6.5 0 0 1 7.4 2.7a6.5 6.5 0 1 0 7.2 9.4z" fill="currentColor"/></svg>
    <svg class="sun" viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="3.4" fill="currentColor"/><g stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.3 4.3l1.4 1.4M14.3 14.3l1.4 1.4M4.3 15.7l1.4-1.4M14.3 5.7l1.4-1.4"/></g></svg>
  </button>`;
}

export function brandMarkSvg() {
  return `<svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M4 10.5 12 5l8 5.5v8a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 18.5v-8Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/><path d="M8 13h8M8 16h8M9 10h6" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>`;
}

export function shieldSvg() {
  return `<svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 3l7 3v5.4c0 4.1-2.8 7.9-7 9.6-4.2-1.7-7-5.5-7-9.6V6l7-3Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M9 12l2 2 4-5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>`;
}

export function faviconSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="gobankcli"><rect width="64" height="64" rx="14" fill="#162032"/><path d="M14 27 32 15l18 12v18a4 4 0 0 1-4 4H18a4 4 0 0 1-4-4V27Z" fill="#eef4ff" stroke="#5ea1ff" stroke-width="3" stroke-linejoin="round"/><path d="M23 32h18M23 39h18M26 25h12" stroke="#1d4ed8" stroke-width="3" stroke-linecap="round"/></svg>`;
}

export function socialCardSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 630" role="img" aria-label="gobankcli social card"><rect width="1200" height="630" fill="#f5f7fb"/><rect x="70" y="70" width="1060" height="490" rx="26" fill="#ffffff" stroke="#dce5eb"/><g transform="translate(110 110) scale(3.2)" color="#1d4ed8">${brandMarkSvg()}</g><text x="110" y="285" font-family="Inter, Arial, sans-serif" font-size="86" font-weight="800" fill="#162032">gobankcli</text><text x="114" y="360" font-family="Inter, Arial, sans-serif" font-size="36" fill="#334155">Local-first, read-only bank transaction archive CLI.</text><text x="114" y="438" font-family="JetBrains Mono, Menlo, monospace" font-size="29" fill="#1d4ed8">SQLite archive | stable CSV | no scraping | no payments</text></svg>`;
}

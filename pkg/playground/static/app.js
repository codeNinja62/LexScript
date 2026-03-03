// LexScript Playground — client-side logic.
//
// - Calls /api/compile and /api/visualize endpoints.
// - Renders Markdown output and FSM (via viz.js) in tabs.

let currentTab = "markdown";
let lastMarkdown = "";
let lastDot = "";

// Keyboard shortcut: Ctrl/Cmd + Enter to compile.
document.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    e.preventDefault();
    compile();
  }
});

// Tab switching.
function showTab(tab) {
  currentTab = tab;
  document.getElementById("tabMarkdown").classList.toggle("active", tab === "markdown");
  document.getElementById("tabFsm").classList.toggle("active", tab === "fsm");

  const output = document.getElementById("output");
  if (tab === "markdown") {
    output.className = "markdown-view";
    output.textContent = lastMarkdown || "Press Compile to see output.";
  } else {
    output.className = "fsm-view";
    if (lastDot) {
      renderDot(lastDot);
    } else {
      output.innerHTML = "<p style='opacity:.5;padding:2em'>Press Compile to generate FSM graph.</p>";
    }
  }
}

// Compile handler — calls both /api/compile and /api/visualize.
async function compile() {
  const source = document.getElementById("editor").value;
  const jurisdiction = document.getElementById("jurisdiction").value;
  const diagList = document.getElementById("diagList");
  const badge = document.getElementById("statusBadge");

  diagList.innerHTML = "<span style='opacity:.5'>Compiling…</span>";
  badge.textContent = "";
  badge.className = "status";

  try {
    // Fire both requests in parallel.
    const [compileRes, vizRes] = await Promise.all([
      fetch("/api/compile", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ source, jurisdiction }),
      }).then((r) => r.json()),
      fetch("/api/visualize", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ source }),
      }).then((r) => r.json()),
    ]);

    // Merge diagnostics.
    const allDiags = [
      ...(compileRes.diagnostics || []),
      ...(vizRes.diagnostics || []).filter(
        (d) => !compileRes.diagnostics.some((c) => c.message === d.message)
      ),
    ];

    // Update diagnostics panel.
    if (allDiags.length === 0) {
      diagList.innerHTML = '<span class="diag-item success">✓ No problems found</span>';
      badge.textContent = "OK";
      badge.className = "status ok";
    } else {
      diagList.innerHTML = allDiags
        .map(
          (d) =>
            `<div class="diag-item">Line ${d.line}:${d.column} — ${escapeHtml(d.message)}</div>`
        )
        .join("");
      badge.textContent = `${allDiags.length} error${allDiags.length > 1 ? "s" : ""}`;
      badge.className = "status err";
    }

    // Update Markdown output.
    lastMarkdown = compileRes.markdown || "";
    lastDot = vizRes.dot || "";

    // Show current tab.
    showTab(currentTab);
  } catch (err) {
    diagList.innerHTML = `<div class="diag-item">Network error: ${escapeHtml(err.message)}</div>`;
    badge.textContent = "Error";
    badge.className = "status err";
  }
}

// Render DOT string using viz.js.
async function renderDot(dot) {
  const output = document.getElementById("output");
  try {
    const viz = await Viz.instance();
    const svg = viz.renderSVGElement(dot);
    output.innerHTML = "";
    output.appendChild(svg);
  } catch (e) {
    output.innerHTML = `<p style="color:var(--error);padding:2em">Render error: ${escapeHtml(e.message)}</p>`;
  }
}

function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

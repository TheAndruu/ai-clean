(async () => {
  const inputEl = document.getElementById("input");
  const outputEl = document.getElementById("output");
  const statsEl = document.getElementById("stats");
  const copyBtn = document.getElementById("copy");
  const loadExampleBtn = document.getElementById("load-example");
  const optNoRejoin = document.getElementById("opt-no-rejoin");
  const optStripAnsi = document.getElementById("opt-strip-ansi");

  // ---------- Install tab switcher ----------
  const tabs = document.querySelectorAll(".terminal-tabs .tab");
  const panes = document.querySelectorAll(".terminal-pane");
  const copyCmdBtn = document.querySelector(".copy-cmd");

  const activatePane = (name) => {
    tabs.forEach(t => t.classList.toggle("active", t.dataset.tab === name));
    panes.forEach(p => p.classList.toggle("active", p.dataset.pane === name));
  };
  tabs.forEach(t => t.addEventListener("click", () => activatePane(t.dataset.tab)));

  if (copyCmdBtn) {
    copyCmdBtn.addEventListener("click", async () => {
      const activePane = document.querySelector(".terminal-pane.active");
      if (!activePane) return;
      const cmdEl = activePane.querySelector(".cmd");
      const cmdText = cmdEl ? cmdEl.textContent : activePane.textContent;
      try {
        await navigator.clipboard.writeText(cmdText.trim());
        const label = copyCmdBtn.querySelector("span");
        const original = label.textContent;
        copyCmdBtn.classList.add("copied");
        label.textContent = "Copied";
        setTimeout(() => {
          copyCmdBtn.classList.remove("copied");
          label.textContent = original;
        }, 1400);
      } catch (err) {
        console.error("copy failed", err);
      }
    });
  }

  // ---------- Cycling CLI names in hero ----------
  const cliCursor = document.querySelector(".cli-cursor");
  if (cliCursor) {
    const names = ["AI CLI", "Claude Code", "Copilot CLI", "Cursor", "Gemini CLI"];
    let idx = 0;
    setInterval(() => {
      idx = (idx + 1) % names.length;
      cliCursor.style.opacity = "0";
      setTimeout(() => {
        cliCursor.textContent = names[idx];
        cliCursor.style.opacity = "1";
      }, 220);
    }, 2400);
    cliCursor.style.transition = "opacity .22s";
  }

  // ---------- WASM demo ----------
  const example = `  │ Here is a small example showing the structure that copilot's CLI │
  │ produces. Each line has left & right border characters, plus  │
  │ trailing whitespace padding so the right border lines up.        │
  │                                                                  │
  │ Code blocks keep their indentation:                              │
  │                                                                  │
  │     def greet(name):                                             │
  │         return f"hello, {name}"                                  │
  │                                                                  │
  │ And prose continues afterwards.                                  │`;

  const setStatus = (msg) => { statsEl.textContent = msg; };
  setStatus("Loading WebAssembly module…");

  const go = new Go();
  let result;
  try {
    if (WebAssembly.instantiateStreaming) {
      result = await WebAssembly.instantiateStreaming(fetch("ai-clean.wasm"), go.importObject);
    } else {
      const resp = await fetch("ai-clean.wasm");
      const bytes = await resp.arrayBuffer();
      result = await WebAssembly.instantiate(bytes, go.importObject);
    }
  } catch (err) {
    setStatus("Failed to load ai-clean.wasm: " + err.message);
    return;
  }
  go.run(result.instance);

  setStatus("");

  const run = () => {
    const text = inputEl.value;
    if (!text) {
      outputEl.value = "";
      copyBtn.disabled = true;
      setStatus("");
      return;
    }
    const out = window.aiClean(text, {
      stripANSI: optStripAnsi.checked,
      noRejoin: optNoRejoin.checked,
    });
    outputEl.value = out.text;
    copyBtn.disabled = !out.text;
    setStatus(formatStats(out.stats));
  };

  const formatStats = (s) => {
    const lines = [];
    if (s.leadingBorderLines)    lines.push(`leading border '${s.leadingBorderChar}' stripped from ${s.leadingBorderLines} line(s)`);
    if (s.trailingBorderLines)   lines.push(`trailing border '${s.trailingBorderChar}' stripped from ${s.trailingBorderLines} line(s)`);
    if (s.boxBorderLinesRemoved) lines.push(`removed ${s.boxBorderLinesRemoved} box-border line(s)`);
    if (s.dedentColumns)         lines.push(`dedented ${s.dedentColumns} column(s) of leading whitespace`);
    if (s.rejoinedLines)         lines.push(`rejoined ${s.rejoinedLines} wrapped line(s)`);
    if (s.blankRunsCollapsed)    lines.push(`collapsed ${s.blankRunsCollapsed} blank-line run(s)`);
    if (s.markdownTableSkipped)  lines.push(`⚠ skipped ${s.markdownTableSkipped} markdown table guard(s) (left '|' borders intact)`);
    if (s.unclosedFence)         lines.push(`⚠ unclosed code fence detected — rejoin disabled inside it`);
    return lines.length ? "ai-clean:\n  " + lines.join("\n  ") : "no changes";
  };

  let pending = 0;
  const schedule = () => {
    clearTimeout(pending);
    pending = setTimeout(run, 80);
  };

  inputEl.addEventListener("input", schedule);
  optNoRejoin.addEventListener("change", run);
  optStripAnsi.addEventListener("change", run);

  loadExampleBtn.addEventListener("click", () => {
    inputEl.value = example;
    run();
    inputEl.focus();
  });

  copyBtn.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(outputEl.value);
      const original = copyBtn.textContent;
      copyBtn.textContent = "Copied";
      setTimeout(() => { copyBtn.textContent = original; }, 1200);
    } catch (err) {
      setStatus("Copy failed: " + err.message);
    }
  });
})();

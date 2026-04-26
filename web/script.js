(async () => {
  const inputEl = document.getElementById("input");
  const outputEl = document.getElementById("output");
  const statsEl = document.getElementById("stats");
  const copyBtn = document.getElementById("copy");
  const loadExampleBtn = document.getElementById("load-example");
  const optNoRejoin = document.getElementById("opt-no-rejoin");
  const optStripAnsi = document.getElementById("opt-strip-ansi");

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

  // Bootstrap the Go runtime.
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

  // Debounce the input so we don't hammer WASM on every keystroke.
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

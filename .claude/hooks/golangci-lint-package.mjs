#!/usr/bin/env node
// PostToolUse hook: run golangci-lint on the package that owns the edited
// *.go file. Per-package (not full repo) keeps latency low while enforcing
// the CLAUDE.md zero-issues contract.
//
// Exit codes:
//   0 — lint clean, or tool missing / file ignored (non-blocking)
//   2 — lint reported issues (surfaced to Claude for remediation)

import { spawnSync } from "node:child_process";
import { existsSync, statSync } from "node:fs";
import { dirname, join } from "node:path";

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => (data += chunk));
    process.stdin.on("end", () => resolve(data));
  });
}

function findModuleRoot(startDir) {
  let dir = startDir;
  while (true) {
    if (existsSync(join(dir, "go.mod"))) return dir;
    const parent = dirname(dir);
    if (parent === dir) return null;
    dir = parent;
  }
}

function hasCommand(cmd) {
  const probe = spawnSync(cmd, ["--version"], { stdio: "ignore" });
  return !probe.error || probe.error.code !== "ENOENT";
}

(async () => {
  try {
    const raw = await readStdin();
    if (!raw.trim()) process.exit(0);

    const payload = JSON.parse(raw);
    const filePath = payload?.tool_input?.file_path;
    if (!filePath || !filePath.endsWith(".go")) process.exit(0);
    if (!existsSync(filePath)) process.exit(0);

    // If golangci-lint is not installed on this machine we silently skip.
    // The CI still enforces it; this hook is just an extra local gate.
    if (!hasCommand("golangci-lint")) process.exit(0);

    const pkgDir = dirname(filePath);
    if (!existsSync(pkgDir) || !statSync(pkgDir).isDirectory()) {
      process.exit(0);
    }

    const modRoot = findModuleRoot(pkgDir);
    if (!modRoot) process.exit(0);

    // Run from module root so golangci-lint picks up .golangci.yml and
    // resolves the package via its directory path.
    const result = spawnSync("golangci-lint", ["run", pkgDir], {
      cwd: modRoot,
      stdio: ["ignore", "inherit", "inherit"],
    });

    if (result.status === 0) process.exit(0);

    // Non-zero: feed diagnostics back to Claude so the next turn can fix.
    process.exit(2);
  } catch (err) {
    console.error("golangci-lint hook error:", err?.message || err);
    process.exit(0);
  }
})();

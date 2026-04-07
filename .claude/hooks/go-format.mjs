#!/usr/bin/env node
// PostToolUse hook: run gofmt (+ goimports if available) on any *.go file
// that Claude just edited or wrote. Silent on success. On gofmt failure,
// exit 2 so Claude sees the diagnostic and can fix it before moving on.

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";

function readStdin() {
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => (data += chunk));
    process.stdin.on("end", () => resolve(data));
  });
}

function hasCommand(cmd) {
  const probe = spawnSync(cmd, ["-h"], { stdio: "ignore" });
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

    // gofmt ships with the Go toolchain — required.
    if (!hasCommand("gofmt")) {
      console.error("go-format: gofmt not on PATH; skipping");
      process.exit(0);
    }

    const gofmt = spawnSync("gofmt", ["-w", filePath], {
      stdio: ["ignore", "inherit", "pipe"],
      encoding: "utf8",
    });
    if (gofmt.status !== 0) {
      process.stderr.write(gofmt.stderr || "gofmt failed\n");
      process.exit(2); // surface to Claude
    }

    // goimports is optional. Try only if it appears installed.
    if (hasCommand("goimports")) {
      const gi = spawnSync("goimports", ["-w", filePath], {
        stdio: ["ignore", "inherit", "pipe"],
        encoding: "utf8",
      });
      if (gi.status !== 0) {
        process.stderr.write(gi.stderr || "goimports failed\n");
        process.exit(2);
      }
    }

    process.exit(0);
  } catch (err) {
    // Never block editing because of a hook bug.
    console.error("go-format hook error:", err?.message || err);
    process.exit(0);
  }
})();

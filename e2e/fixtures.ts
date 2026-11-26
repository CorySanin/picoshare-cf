import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { once } from "node:events";
import { mkdtemp, rm } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { setTimeout as delay } from "node:timers/promises";
import { test as base, expect } from "@playwright/test";

const serverStartupTimeoutMs = 10_000;
const serverShutdownTimeoutMs = 20_000;

async function getAvailablePort(): Promise<number> {
  const server = createServer();

  await new Promise<void>((resolveListen, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolveListen);
  });

  const address = server.address();
  if (address === null || typeof address === "string") {
    server.close();
    throw new Error("failed to determine an available TCP port");
  }

  await new Promise<void>((resolveClose, reject) => {
    server.close((error) => {
      if (error) {
        reject(error);
        return;
      }
      resolveClose();
    });
  });
  return address.port;
}

async function waitForServer(
  baseURL: string,
  child: ChildProcessWithoutNullStreams,
  getDiagnostics: () => string,
): Promise<void> {
  const deadline = Date.now() + serverStartupTimeoutMs;
  let lastRequestError = "no request attempted";

  while (Date.now() < deadline) {
    if (child.exitCode !== null) {
      throw new Error(
        `PicoShare exited with code ${child.exitCode} during startup.\n${getDiagnostics()}`,
      );
    }

    try {
      const response = await fetch(baseURL, { redirect: "manual" });
      if (response.status === 200) {
        return;
      }
      lastRequestError = `unexpected HTTP status ${response.status}`;
    } catch (error) {
      lastRequestError = String(error);
    }
    await delay(50);
  }

  throw new Error(
    `PicoShare did not start at ${baseURL}: ${lastRequestError}.\n${getDiagnostics()}`,
  );
}

async function stopServer(
  child: ChildProcessWithoutNullStreams,
): Promise<void> {
  if (child.pid === undefined) {
    throw new Error("PicoShare process did not start");
  }
  if (child.exitCode !== null || child.signalCode !== null) {
    throw new Error(
      `PicoShare exited before teardown with code ${child.exitCode} and signal ${child.signalCode}`,
    );
  }

  const exit = once(child, "exit");
  child.kill("SIGTERM");
  const stopped = await Promise.race([
    exit.then(([code, signal]) => ({ code, signal })),
    delay(serverShutdownTimeoutMs).then(() => undefined),
  ]);
  if (stopped !== undefined) {
    if (stopped.code !== 0 || stopped.signal !== null) {
      throw new Error(
        `PicoShare stopped with code ${stopped.code} and signal ${stopped.signal}`,
      );
    }
    return;
  }

  child.kill("SIGKILL");
  await exit;
  throw new Error(
    `PicoShare did not stop within ${serverShutdownTimeoutMs} ms and required SIGKILL`,
  );
}

export const test = base.extend({
  // This fixture owns PicoShare's complete lifecycle for one test. A unique
  // process, port, temporary directory, and SQLite file prevent parallel tests
  // from sharing application or database state. The fixture waits for a real
  // HTTP response before handing the URL to Playwright, then records process
  // output and removes all temporary state after the test finishes.
  baseURL: async ({}, use, testInfo) => {
    const port = await getAvailablePort();
    const testDirectory = await mkdtemp(join(tmpdir(), "picoshare-e2e-"));
    const dbPath = join(testDirectory, "store.db");
    const baseURL = `http://127.0.0.1:${port}`;
    const binary = resolve("bin/picoshare-dev");
    const child = spawn(binary, ["-db", dbPath], {
      cwd: testDirectory,
      env: {
        ...process.env,
        PORT: String(port),
        LITESTREAM_BUCKET: "",
        PS_BEHIND_PROXY: "",
        PS_SHARED_SECRET: "dummypass",
        PS_SHARED_SECRET_FILE: "",
        TZ: "UTC",
      },
      stdio: "pipe",
    });

    let output = "";
    child.stdout.on("data", (data: Buffer) => {
      output += data.toString();
    });
    child.stderr.on("data", (data: Buffer) => {
      output += data.toString();
    });
    let spawnError: Error | undefined;
    child.once("error", (error) => {
      spawnError = error;
    });

    const getDiagnostics = (): string => {
      const error = spawnError ? `spawn error: ${spawnError.message}\n` : "";
      return `${error}PicoShare output:\n${output || "<no output>"}`;
    };

    try {
      await waitForServer(baseURL, child, getDiagnostics);
      await use(baseURL);
    } finally {
      try {
        await stopServer(child);
      } finally {
        try {
          await testInfo.attach("picoshare.log", {
            body: getDiagnostics(),
            contentType: "text/plain",
          });
        } finally {
          await rm(testDirectory, { recursive: true, force: true });
        }
      }
    }
  },
});

export { expect };
export type { Page } from "@playwright/test";

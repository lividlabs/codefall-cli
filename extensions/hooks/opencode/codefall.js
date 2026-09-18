// codefall's OpenCode plugin: the same hooks the JSON harnesses get, in OpenCode's plugin
// form. Bash calls pass through the shared merge guard, and session starts get primed with
// what Beads knows and told what needs attention in the project.
//
// The guard and the notice are delegated to the scripts in hooks/shared/ — one logic source
// for every harness; this file is only the adapter from OpenCode's shapes to them.
//
// Both are found from this file's own directory, not from the worktree root: init installs
// the plugin at <install>/.opencode/plugins/ and the scripts at <install>/.codefall/, and the
// install is the repository root only when init was run there.
export const CodefallPlugin = async ({ client, $ }) => {
  const guard = `${import.meta.dir}/../../.codefall/hooks/shared/codefall-block-merge-to-main.sh`;
  const notice = `${import.meta.dir}/../../.codefall/hooks/shared/codefall-session-notice.sh`;

  return {
    // Deny merges and pushes to the default branch. The script answers exit 2 for an
    // objection; OpenCode blocks the tool call on a thrown error. The payload goes in through
    // piped stdin with no shell in between, because the command it carries is model-controlled.
    "tool.execute.before": async (input, output) => {
      if (input.tool !== "bash") return;

      // The script speaks the Claude-shaped payload, so the shape is rebuilt here
      // rather than read from output directly.
      const payload = JSON.stringify({ tool_input: { command: output.args.command } });

      // Missing script means the extension is not installed here (or the project moved); the
      // guard stands down rather than throwing on every Bash call. Fail-open is the deliberate
      // trade: this is one layer, and a repository ruleset protecting the default branch is the
      // backstop — the shared script states the same limit about itself.
      if (!(await Bun.file(guard).exists())) return;

      const proc = Bun.spawn(["bash", guard], {
        stdin: new TextEncoder().encode(payload),
        stdout: "ignore",
        stderr: "pipe",
      });
      const [stderr, code] = await Promise.all([new Response(proc.stderr).text(), proc.exited]);
      if (code !== 0) {
        throw new Error(`codefall: ${stderr.trim() || "the merge guard denied this command."}`);
      }
    },

    // Prime a fresh session with what Beads knows and with what the project needs doing:
    // bd's report and codefall's notice go in as context without triggering a response.
    // Anything missing (bd, the script, the session) is skipped — a session hook must never
    // break the session that starts it.
    event: async ({ event }) => {
      if (event.type !== "session.created") return;

      const id = event.properties?.id;
      if (!id) return;

      const parts = [];

      try {
        const prime = await $`bd prime`.nothrow().quiet();
        if (prime.exitCode === 0) {
          const text = prime.stdout.toString().trim();
          if (text) parts.push({ type: "text", text });
        }
      } catch {
        // A failed prime is a missed convenience, not an error worth a session over.
      }

      // The notice says what needs attention — the checkout behind the default branch, an
      // environment that has not been refreshed, a testing root nobody declared, a Beads
      // precondition — and prints nothing when the project is current. Plain text, because
      // that is what goes into a prompt part; --hook-json is for the harnesses whose
      // SessionStart entry reads JSON. It reports only: nothing here pulls or updates.
      try {
        if (await Bun.file(notice).exists()) {
          const reported = Bun.spawn(["bash", notice], { stdout: "pipe", stderr: "ignore" });
          const [said] = await Promise.all([new Response(reported.stdout).text(), reported.exited]);
          const text = said.trim();
          if (text) parts.push({ type: "text", text });
        }
      } catch {
        // The same trade the prime makes: the session opens either way.
      }

      if (parts.length === 0) return;

      try {
        await client.session.prompt({ path: { id }, body: { noReply: true, parts } });
      } catch {
        // A session that will not take the context still starts, which is the point.
      }
    },
  };
};

// codefall's OpenCode plugin: the same two hooks the JSON harnesses get, in OpenCode's
// plugin form. Bash calls pass through the shared merge guard, session starts get primed
// with what Beads knows.
//
// The guard is delegated to hooks/shared/codefall-block-merge-to-main.sh — one logic
// source for every harness; this file is only the adapter from OpenCode's shapes to it.
export const CodefallPlugin = async ({ client, $, directory, worktree }) => {
  const root = worktree || directory;
  const guard = `${root}/.agents/hooks/shared/codefall-block-merge-to-main.sh`;

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

    // Prime a fresh session with what Beads knows: bd's report goes in as context
    // without triggering a response. Anything missing (bd, the session) is skipped —
    // a session hook must never break the session that starts it.
    event: async ({ event }) => {
      if (event.type !== "session.created") return;

      const id = event.properties?.id;
      if (!id) return;

      try {
        const prime = await $`bd prime`.nothrow().quiet();
        if (prime.exitCode !== 0) return;

        const text = prime.stdout.toString().trim();
        if (!text) return;

        await client.session.prompt({
          path: { id },
          body: { noReply: true, parts: [{ type: "text", text }] },
        });
      } catch {
        // A failed prime is a missed convenience, not an error worth a session over.
      }
    },
  };
};

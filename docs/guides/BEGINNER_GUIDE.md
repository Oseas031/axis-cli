# Axis Beginner Guide: Getting Started with axis-up

**[Chinese version / 中文版](../zh/guides/BEGINNER_GUIDE.md)**

This guide is for first-time Axis users. You don't need to understand Axis's architecture, and you don't need a real model API key. Just use `axis-up` to complete environment checks, build Axis, run a demo, and understand what to do next.

**Important**: `axis-up` is an external onboarding tool. It's recommended for first-time setup, but for long-term use you should learn the Axis CLI directly. `axis-up` does not modify Axis source code, does not import Axis internal packages, and only interacts with Axis through public commands.

---

## 0. Axis vs axis-up: A First-Principles Comparison

Understanding the difference between Axis and axis-up starts from their design goals and target audiences:

| Dimension | Axis | axis-up |
|-----------|------|---------|
| **Primary design goal** | Agent autogenesis execution substrate | Human quick-start assistant |
| **Target audience** | AI Agents | Human users |
| **Core responsibility** | Task scheduling, contract execution, state management | Environment check, build guidance, zero-config demo |
| **Long-term positioning** | The system itself | Temporary bridge |
| **Interaction** | CLI / shell / public API | CLI / command-line tool |
| **Dependencies** | Runs independently | Depends on Axis public CLI |

**First-principles conclusion**:

- **Axis exists to let Agents express work as tasks, obtain context, execute actions, validate results, and earn greater autonomy through reliable performance.**
- **axis-up exists to prevent first-time human users from getting stuck on environment checks, builds, and provider configuration.**

Therefore, axis-up is the ladder to help humans "get on board"; Axis is the vehicle itself. Once aboard, you should learn to drive Axis directly rather than always relying on the ladder.

---

## 1. What Will You Do With axis-up?

For first-time setup, just remember four commands:

```powershell
.\axis-up.exe start
.\axis-up.exe check
.\axis-up.exe demo
.\axis-up.exe fix
```

They correspond to four user intents:

- **`start`**: I want to get Axis running for the first time.
- **`check`**: I want to know if the current environment is ready.
- **`demo`**: I want to see a minimal runnable demonstration.
- **`fix`**: I want to fix common onboarding issues.

If you don't know what to run, start with:

```powershell
.\axis-up.exe start
```

---

## 2. Prerequisites: Confirm Go is Available

Both Axis and `axis-up` are Go programs, so you need Go installed first.

In PowerShell:

```powershell
go version
```

If you see output like this, Go is available:

```text
go version go1.xx.x windows/amd64
```

If you don't see a version number, install Go first, then continue.

---

## 3. Build axis-up

Navigate to the `axis-up` tool directory:

```powershell
cd c:\Users\ASUS\Desktop\axis-cli\tools\axis-up
```

Build `axis-up.exe`:

```powershell
go build -o axis-up.exe .
```

After successful build, the current directory will contain:

```text
axis-up.exe
```

This is the external onboarding entry point for new users.

---

## 4. First Launch: axis-up start

Run:

```powershell
.\axis-up.exe start
```

`start` automatically determines what you need:

```text
Detect Axis repo → Detect Go → Build axis-dev.exe if needed → Use mock provider → Run demo
```

You'll see output like:

```text
Start Axis
----------
Goal:
  Get a first Axis run working without changing Axis core behavior.

[ok] Axis repo: C:\Users\ASUS\Desktop\axis-cli
[ok] Go: go version go1.xx.x windows/amd64
[ok] Axis binary: C:\Users\ASUS\Desktop\axis-cli\axis-dev.exe
```

If `axis-dev.exe` is missing, `axis-up start` will explain why it needs to be built and run something like:

```text
go build -o axis-dev.exe ./cmd/axis
```

> Note: Axis has completed all milestones M1-M6, including real LLM integration, sandboxed evolution, self-judgement, and more. This guide focuses on the first-time experience.

Using `axis-dev.exe` on Windows avoids overwriting an existing `axis.exe` in the project root.

---

## 5. Why Mock Provider by Default?

The first-time experience should not get blocked by:

- API keys
- External network access
- Model billing
- Provider compatibility

So `axis-up` defaults to mock provider.

The mock provider's goal is not to demonstrate real model capabilities, but to let you see Axis's minimal closed loop:

```text
Submit task → Schedule task → Execute default contract → Update task status
```

Once you've confirmed Axis works end-to-end, configuring a real provider is more reliable.

---

## 6. Check Environment: axis-up check

If you're unsure about the current state, run:

```powershell
.\axis-up.exe check
```

It checks:

- Whether the current directory is an Axis project
- Whether Go is available
- Whether the Axis binary exists
- Whether provider configuration exists
- What you should do next

Example output:

```text
Axis readiness check
--------------------
[ok] Axis repo: C:\Users\ASUS\Desktop\axis-cli
[ok] Go: go version go1.xx.x windows/amd64
[ok] Axis binary: C:\Users\ASUS\Desktop\axis-cli\axis-dev.exe
[ok] Provider config: not required for mock provider

Next:
  Run: axis-up demo
```

Key point: `axis-up` doesn't just tell you "it failed"—it tells you what to do next.

---

## 7. Run Demo Separately: axis-up demo

If the environment is already prepared, you can run:

```powershell
.\axis-up.exe demo
```

It submits a demo task through the Axis public CLI:

```text
axis-up-demo
```

You might see:

```text
Axis demo
---------
What this does:
  Submit one demo task through the public Axis CLI using the mock provider.

Action:
  axis run axis-up-demo --provider mock

Task axis-up-demo submitted successfully
```

Seeing this line means Axis can accept tasks:

```text
Task axis-up-demo submitted successfully
```

---

## 8. Fix Common Issues: axis-up fix

If `check` reports a missing binary, or `demo` fails to run, try:

```powershell
.\axis-up.exe fix
```

`fix` only performs safe repairs, for example:

- Missing `axis-dev.exe`: build it with Go.
- Provider config doesn't exist: explain that mock provider needs no config.
- Wrong directory: suggest running inside the Axis project.

`fix` will NOT:

- Silently modify Axis source code
- Silently overwrite provider configuration
- Delete your files
- Configure a real model API key for you

---

## 9. What Does axis-up Call Behind the Scenes?

`axis-up` is not a new Axis core. It just calls the public Axis CLI for you.

For example, `axis-up demo` is similar to running:

```powershell
cd c:\Users\ASUS\Desktop\axis-cli
.\axis-dev.exe --provider mock run axis-up-demo
```

In other words:

- `axis-up` handles beginner guidance
- `axis-dev.exe` actually executes Axis commands
- Mock provider enables zero-config demos

This is why `axis-up` can help beginners without intruding into Axis core.

---

## 10. Next Level: Enter Axis Shell

After running `axis-up start` or `axis-up demo` successfully, try Axis's interactive shell.

Go back to the Axis project root:

```powershell
cd c:\Users\ASUS\Desktop\axis-cli
```

Start the shell:

```powershell
.\axis-dev.exe --provider mock shell
```

You'll see:

```text
Axis shell started. Type 'help' for commands, 'exit' to quit.
axis>
```

In the shell, type:

```text
help
run demo-task
status demo-task
exit
```

This step is not required for first-time setup, but it helps you understand Axis's real interaction mode.

---

## 11. FAQ

### Q1: Do I need to configure a real model first?

No.

First-time setup defaults to mock provider. You don't need an API key or network access.

### Q2: Will axis-up modify Axis itself?

No.

`axis-up`'s boundaries are:

- Can check environment
- Can build `axis-dev.exe`
- Can call Axis public commands
- Can explain next steps

It does not import Axis internal packages, nor modify Axis source code.

### Q3: Why no web UI?

Axis core follows `bash is all you need`, prioritizing CLI. But if you need a visual interface, use [axis-gui](../../tools/axis-gui/)—a local Web Dashboard connected to the Local Control Plane.

### Q4: `status` says task not found?

Cross-command submit and query requires starting the local runtime first:

```powershell
# Terminal A
.\axis-dev.exe start

# Terminal B
.\axis-dev.exe ask "demo task" --submit --task-id demo
.\axis-dev.exe status demo
```

Or operate within the same shell session:

```powershell
.\axis-dev.exe --provider mock shell
```

### Q5: What documentation should I read next?

Recommended order:

1. `README.md`: Project overview and CLI command reference
2. `docs/guides/QUICKSTART.md`: Developer quick start
3. `docs/architecture/agent-native-first-principles.md`: **Read before coding**
4. `docs/architecture/bash-is-all-you-need.md`: Understand interaction principles
5. `docs/product/ROADMAP.md`: Milestone roadmap

---

## 12. Shortest Path Summary

To get Axis running as fast as possible:

```powershell
cd c:\Users\ASUS\Desktop\axis-cli\tools\axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

To check status later:

```powershell
.\axis-up.exe check
```

To re-run the demo:

```powershell
.\axis-up.exe demo
```

To fix common issues:

```powershell
.\axis-up.exe fix
```

This is the complete path for getting started with Axis using `axis-up`.

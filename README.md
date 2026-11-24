# LiteCode

A light, keyboard-first IDE with dockable panels and a built-in Git workflow,
built on [Giode](https://github.com/xmhrnx/giode). No JVM, no indexer: a
native binary that uses no CPU while idle.

- `docs/FEATURES.md`: the feature research and what LiteCode keeps.

## Run

```bash
go run . [project-folder]
```

With no folder it reopens the last project, or shows the welcome screen.

## Keymap (defaults; Mod = Ctrl, Cmd on macOS)

| Keys | Action |
| --- | --- |
| Shift Shift | Search Anything |
| Mod+Shift+N / Mod+Shift+A | Go to File / Run Action |
| Mod+Shift+F | Find in Files |
| Mod+E | Recent Files |
| Mod+G | Go to Line |
| Ctrl+Space | Show completions (they also open as you type) |
| Alt+1 / Alt+0 / Alt+9 / Alt+F12 | Files / Changes / History / Terminal panels |
| Shift+Esc / Mod+Shift+F12 | Hide panel / Toggle all panels |
| Mod+K / Mod+Shift+K / Mod+T | Commit / Push / Pull |
| Mod+Shift+` | Branches |
| Mod+Alt+Z | Discard changes |
| F7 / Shift+F7 | Next / previous change in a diff |
| Mod+D / Mod+Y / Mod+/ | Duplicate / delete / comment line |
| Alt+Shift+Up/Down | Move line |
| Mod+Alt+L / Mod+S | Format / Save all (formats on save) |
| Ctrl+Tab, Alt+Left/Right, Mod+F4 | Switch / close tabs |
| Mod+Alt+S | Settings |
| Alt+I | Assistant panel |

Change any of them in Settings › Keymap (click an action, press the keys).
Overrides are saved to `<config>/litecode/keymap.json`.

## Panels

Right-click a panel's button in the activity bar (or use its ⋮ menu) › Move To
› Left, Bottom or Right. Drag the edge between a panel and the editor to
resize. The arrangement is global, so a terminal moved to the right stays
there in every project.

## Formatting

Settings › On Save is global: Prettier runs for JS/TS/CSS/HTML/JSON/
Markdown/YAML in every project (the project's `node_modules/.bin/prettier`,
else the one on PATH), and other languages use their formatter (gofmt,
rustfmt, or a plugin's).

## AI Assistant

The Assistant panel (Alt+I, docked right) chats about your code using a model you
connect in Settings › AI Assistant. The open file goes with your message while its
chip is on (files over 256 KB are refused, not cut down).

| Kind | Providers | Wire format |
| --- | --- | --- |
| Cloud (your own key) | Anthropic, OpenAI, Google Gemini, Mistral, Groq, OpenRouter, DeepSeek, xAI, Together AI, Fireworks AI, Azure OpenAI | Anthropic through its official Go SDK; Gemini through its native `streamGenerateContent`; the rest through Chat Completions |
| Local (no key) | Ollama (native `/api/chat`), LM Studio, llama.cpp server, vLLM, Jan, LocalAI | Chat Completions, except Ollama |
| Custom | anything you add: name, API (OpenAI-, Anthropic-, Gemini- or Ollama-style), base URL, auth header and prefix, extra headers, model, output-limit field | |

- **Keys** are stored in the system keychain (Keychain, Credential Manager/DPAPI,
  Secret Service), never in config files. If none is saved, the provider's usual
  environment variable is used (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, …).
- **Models:** each provider has suggested defaults (checked against its docs in
  November 2026). Fetch Models pulls the live list from the provider.
- **Test Connection** sends a one-word prompt and shows the reply and timing.
- **Claude:** the default is `claude-opus-5-5`. On Anthropic's own API, requests opt into
  server-side refusal fallbacks (beta `server-side-fallback-2026-07-01`,
  `fallbacks: "default"`), so a declined request is retried on a fallback model in the
  same call. Custom Anthropic-compatible gateways don't get that flag.
- Non-secret settings live in `<config>/litecode/ai.json`.

## Plugins

A plugin is a folder in `<config>/litecode/plugins/` with a `plugin.json`
describing languages: extensions, file names, comment tokens, indentation, a
formatter command and a language server. See `docs/plugin-example/zig`.
`<config>` is `~/.config` on Linux, `%AppData%` on Windows,
`~/Library/Application Support` on macOS.

## Layout

```
main.go          wiring only
assets/          Inter + JetBrains Mono fonts (SIL OFL), Lucide icons (ISC)
internal/        logic, no Giode imports, unit-tested
  git/ diff/ search/ keymap/ lang/ format/ layout/ recent/ workspace/ textops/ project/ config/
ui/              one component per file
  app/ core/ frame/ toolwin/ editor/ projectview/ gitui/ search/ settings/ welcome/ terminal/ widgets/ theme/
```

## License

LiteCode is licensed under the [Apache License 2.0](LICENSE). Bundled fonts (SIL OFL 1.1) and
icons (ISC) keep their own licenses; see [NOTICE](NOTICE).

# FreeRouter 🚀

> Run OpenRouter **free** models like Ollama — Cursor-compatible local proxy.  
> Built with Go + [Cobra](https://github.com/spf13/cobra).

`fr` starts a local HTTP server that speaks the OpenAI Chat API so
Cursor (and any OpenAI-compatible client) can use free models from OpenRouter
without any cost.

---

## Install

### Prerequisites

- Go 1.23+
- A free [OpenRouter](https://openrouter.ai) account + API key

```bash
# Option A — go install (recommended)
go install github.com/freerouter/freerouter@latest

# Option B — build from source
git clone https://github.com/freerouter/freerouter
cd freerouter
make install        # runs `go install`, adds to $GOPATH/bin
```

---

## Setup — API Key

```bash
# Option A — environment variable (add to ~/.bashrc / ~/.zshrc)
export OPENROUTER_API_KEY=sk-or-...

# Option B — ~/.freerouter file
echo "OPENROUTER_API_KEY=sk-or-..." > ~/.freerouter

# Option C — .env in your project directory
echo "OPENROUTER_API_KEY=sk-or-..." > .env

# Option D — pass it directly
freerouter run llama-4-scout --key sk-or-...
```

---

## Commands

### `freerouter models`

List all free models on OpenRouter, grouped by provider.

```bash
freerouter models
freerouter models --search llama    # filter by name / ID
freerouter models --json            # raw JSON output
```

**Example output:**
```
✨ 42 free model(s) on OpenRouter

  google
    google/gemma-3-27b-it:free       [131k ctx]
    google/gemma-3-4b-it:free        [131k ctx]

  meta-llama
    meta-llama/llama-4-maverick:free [1M ctx]
    meta-llama/llama-4-scout:free    [512k ctx]

  ...

  Run: freerouter run <model-id> to start a local proxy
```

---

### `freerouter run <model>`

Start a local **OpenAI-compatible proxy** for the chosen model.

```bash
freerouter run meta-llama/llama-4-scout:free   # full ID
freerouter run llama-4-scout                   # fuzzy match
freerouter run scout                           # even shorter
freerouter run deepseek-r1 --port 8080         # custom port
```

**Flags:**

| Flag     | Short | Default | Description                               |
| -------- | ----- | ------- | ----------------------------------------- |
| `--port` | `-p`  | `11434` | TCP port to bind (same default as Ollama) |
| `--key`  | `-k`  | *(env)* | OpenRouter API key                        |

**Example output:**
```
✔ Model resolved: meta-llama/llama-4-scout:free

  🟢 freerouter is running!

  Model  meta-llama/llama-4-scout:free
  URL    http://localhost:11434
  Ctx    512k tokens

  ── Add to Cursor ──────────────────────────────────────
  Cursor → Settings → Models → Add Custom Model:

    Base URL : http://localhost:11434/v1
    Model    : meta-llama/llama-4-scout:free
    API Key  : freerouter

  Press Ctrl+C to stop.
```

---

## Adding to Cursor

1. Start `freerouter run <model>` in a terminal
2. Open **Cursor** → `Settings` → **Models** → **Add Custom Model**
3. Fill in:

   | Field    | Value                          |
   | -------- | ------------------------------ |
   | Base URL | `http://localhost:11434/v1`    |
   | Model    | *(model ID shown in terminal)* |
   | API Key  | `freerouter` *(any string)*    |

4. Select your new model from the model picker — done!

---

## Quick curl test

```bash
curl http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-llama/llama-4-scout:free",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": false
  }'
```

Streaming works too — set `"stream": true`.

---

## Running multiple models

Run `freerouter run` on different ports and add each to Cursor:

```bash
freerouter run llama-4-scout   --port 11434   # terminal 1
freerouter run deepseek-r1     --port 11435   # terminal 2
freerouter run gemma-3-27b-it  --port 11436   # terminal 3
```

---

## How it works

```
Cursor / Editor
      │  OpenAI Chat API  (localhost:11434)
      ▼
 freerouter proxy  (Go HTTP server)
      │  OpenRouter REST API
      ▼
 Free model  (Llama 4, DeepSeek R1, Gemma 3, …)
```

- `models` → `GET https://openrouter.ai/api/v1/models`, filter `pricing.prompt == "0"`
- `run` → binds a local `net/http` server; every `/v1/chat/completions` POST
  is forwarded to OpenRouter with your key injected; SSE chunks are flushed
  immediately so streaming feels instant in Cursor.

---

## Best free coding models (2025)

| Model ID                           | Context | Strengths           |
| ---------------------------------- | ------- | ------------------- |
| `meta-llama/llama-4-scout:free`    | 512k    | Fast, large context |
| `meta-llama/llama-4-maverick:free` | 1M      | Strong reasoning    |
| `google/gemma-3-27b-it:free`       | 131k    | Solid all-rounder   |
| `deepseek/deepseek-r1:free`        | 164k    | Best free reasoning |

---

## License

GPL-3.0

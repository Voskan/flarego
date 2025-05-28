# FlareGo 🔥

> **Live scheduler flame‑visualiser for Go 1.19+** – sample goroutines, GC, heap, blocked states in real time and explore the data through an embeddable React dashboard.

[![CI](https://github.com/Voskan/flarego/actions/workflows/ci.yml/badge.svg)](https://github.com/Voskan/flarego/actions/workflows/ci.yml)
[![Release](https://github.com/Voskan/flarego/actions/workflows/release.yml/badge.svg)](https://github.com/Voskan/flarego/actions/workflows/release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/Voskan/flarego)](https://goreportcard.com/report/github.com/Voskan/flarego)

FlareGo consists of three high‑level pieces:

| Component     | Purpose                                                                                                                               | Binary / Image                                |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| **Agent**     | In‑process collector that samples goroutines, heap, GC, blocked counts and ships flamegraph snapshots.                                | `flarego-agent` / `ghcr.io/flarego/agent`     |
| **Gateway**   | gRPC fan‑out service that receives snapshots, keeps an in‑memory/Redis retention ring and broadcasts to WebSocket & gRPC‑web clients. | `flarego-gateway` / `ghcr.io/flarego/gateway` |
| **Dashboard** | React + Vite SPA (usable as standalone page or IDE/WebView component).                                                                | served from the gateway under `/`             |

---

## ✨ Features

- **Real‑time flamegraphs** (≤ 20 ms latency)
- Goroutine, GC, heap, blocked **samplers**
- **Before / After diff** & desert‑sunset palette 🌄
- JSON or protobuf **chunk encoding** (2–3× smaller than pprof)
- Alert **DSL** with Slack / Webhook / Jira sinks
- Prometheus metrics & OTLP span correlation
- Zero‑dependency **agent attach** (`go:linkname` free)

---

## 🚀 Quick‑start (Docker Compose)

```bash
# Clone repo & start gateway + demo app
$ git clone https://github.com/Voskan/flarego && cd flarego
$ docker compose -f deployments/docker-compose.yaml up -d

# Point your browser at http://localhost:8080
```

The demo runs `examples/basic` inside the compose network; the agent streams to `gateway:4317` and the dashboard auto‑connects.

---

## 🛠️ Local build

```bash
# Gateway (static binary)
go build -o bin/flarego-gateway ./cmd/flarego-gateway

# Agent (self‑sampling demo)
go build -o bin/flarego-agent ./cmd/flarego-agent

# Web dashboard (hot‑reload)
cd web && npm i && npm run dev
```

---

## 🧩 Attach to a running binary

```bash
# Compile your app with -buildmode=pie or default (Go ≥1.20)
# Then in another terminal:
$ flarego attach --pid <PID> --gateway localhost:4317
```

`attach` injects a tiny sampler via `runtime.StartTrace` without requiring cgo or `dlopen`.

---

## 📟 CLI reference

| Command          | Description                                     |
| ---------------- | ----------------------------------------------- |
| `flarego record` | Capture flamegraphs to `my.fgo` (gzipped JSON). |
| `flarego replay` | Pretty‑print or stream a recorded file.         |
| `flarego diff`   | (coming) side‑by‑side diff of two `.fgo` files. |

See `docs/cli-reference.md` for exhaustive flags.

---

## ☁️ Kubernetes

Minimal manifests are provided under `deployments/kubernetes/` – suitable for dev clusters:

```bash
kubectl apply -f deployments/kubernetes/gateway-deployment.yaml
kubectl apply -f deployments/kubernetes/agent-daemonset.yaml
```

Add `FLAREGO_GW_LISTEN=:4317` to your ConfigMap/Secret to customise.

---

## 🖥️ Architecture overview

```
  +-----------+        gRPC           +-----------+    WebSocket/gRPC‑web   +---------+
  |  Agent(s) |  ──────────────────▶ |  Gateway  | ──────────────────────▶ |  UI SPA |
  +-----------+      Flamegraph       +-----------+        JSON             +---------+
          ▲                              │  Redis /        ▲
          │                              │  In‑mem         │ Alerts
          │         Control / JWT        ▼                 │
          └─────────────────────────── HTTP REST ◀─────────┘
```

Detailed docs: [docs/architecture.md](docs/architecture.md) & [docs/dataflow.md](docs/dataflow.md).

---

## 🥳 Contributing

1. Fork + clone
2. `make dev` (runs linters & UI watcher)
3. Send PR – ensure `make lint && make test` passes

All code is licensed under Apache‑2.0. By contributing you agree to that licence.

---

## 📝 License

```
Apache License, Version 2.0
Copyright (c) 2025 FlareGo authors
```

# LibGen Downloader (Native Go + Fyne GUI)

[![CI](https://github.com/anilpdv/libgen-gui/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/anilpdv/libgen-gui/actions/workflows/ci.yml)
[![Pages](https://github.com/anilpdv/libgen-gui/actions/workflows/pages.yml/badge.svg?branch=main)](https://anilpdv.github.io/libgen-gui)
[![Go Report Card](https://goreportcard.com/badge/github.com/anilpdv/libgen-gui)](https://goreportcard.com/report/github.com/anilpdv/libgen-gui)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/anilpdv/libgen-gui?color=blue)](https://github.com/anilpdv/libgen-gui/releases)

> 🌐 **Official Website & Documentation**: [https://anilpdv.github.io/libgen-gui](https://anilpdv.github.io/libgen-gui)

A fast, lightweight, native cross-platform desktop and mobile app for searching, discovering, and downloading books and research papers from Library Genesis mirrors.

Built with **Go** and the **Fyne v2** vector UI toolkit for 100% native performance with zero webview or Electron overhead.

---

<p align="center">
  <img src="resources/libgen_downloader_preview.png" alt="LibGen Downloader Preview (macOS & Android)" width="850" style="border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);" />
</p>

---

## 📱 Platform Verification & Support

| Platform | Verification Status | Target Package | Download |
|---|:---:|:---:|---|
| **macOS (Apple Silicon & Intel)** | ✅ **Primary Verified & Tested** | Native `.app` / `.zip` | [Download macOS Build (v2.0.0)](https://github.com/anilpdv/libgen-gui/releases) |
| **Android (API 26–35)** | ✅ **Primary Verified & Tested** | Signed `.apk` (SAF support) | [Download Android APK (v2.0.0)](https://github.com/anilpdv/libgen-gui/releases) |
| **Linux (X11 / Wayland)** | 🔄 Verified & Cross-compilation ready | Native binary / `.tar.gz` | [Build from Source](#building--running) |
| **Windows (10 / 11)** | 🔄 Verified & Cross-compilation ready | Native `.exe` | [Build from Source](#building--running) |

---

## ✨ Key Features (v2.0.0)

- **Decoupled Architecture**: Full clean-architecture separation (`internal/network`, `internal/storage`, `internal/search`, `internal/download`, `internal/settings`, `internal/app`).
- **Configurable Download Destinations**: 
  - Desktop: choose any writable local directory.
  - Android: choose a folder through the Storage Access Framework (SAF).
  - Existing queued downloads retain their original destination captured at enqueue time.
  - New downloads automatically use the newly selected destination.
  - Selected destination persists across application restarts.
- **Multi-Mirror Automatic Failover**: Continuous background latency probing and automatic failover across official mirrors (`libgen.is`, `libgen.rs`, `libgen.st`, `libgen.li`).
- **Atomic Resumable Downloads**: Temporary `.part` file isolation with HTTP Range chunk resumption and atomic rename upon completion.
- **Crash-Safe Queue Persistence**: Thread-safe serialized FIFO download queue backed by an atomic `.tmp` JSON ledger.
- **Android Storage Access Framework (SAF)**: Native `DocumentFile` scoped storage and custom folder tree URI permissions for Android 10 through 15.
- **Debounced Instant Search**: Multi-field search across title, author, series, publisher, year, ISBN, and MD5 with instant format badges.
- **Zero Webview Overhead**: Compiles to a single static binary consuming less than 35MB RAM, powered by OpenGL/Metal hardware acceleration.

### Download Destination Behavior
Changing the download location does not move active or queued downloads. Existing tasks continue using the destination captured when they were added. New tasks use the newly selected location.

### Troubleshooting
#### Android says folder permission was lost
Android can revoke access to a previously selected document tree. Open **Settings → Downloads → Change Folder** and select the folder again.

---

## 🚀 Installation & Releases

### 📥 Download Pre-built Releases
Grab ready-to-run releases from the [GitHub Releases](https://github.com/anilpdv/libgen-gui/releases) page:
- **macOS**: `LibGen.Downloader-v2.0.0-macos.zip` (unzip and run `LibGen Downloader.app`)
- **Android**: `LibGen.Downloader-v2.0.0-android.apk` (install on Android 8.0+)
- **Desktop Binary**: `libgen-gui-v2.0.0`

---

## 🏗️ Architecture Overview

```
libgen-gui/
├── cmd/
│   └── libgen-gui/        # Entry point binary
│       └── main.go
├── internal/
│   ├── app/               # Composition root and wiring
│   ├── download/          # Download manager, state machine, queue & ledger
│   ├── network/           # Resilient HTTP client, mirrors, errors & backoff
│   ├── search/            # Domain models, search controller & service
│   ├── settings/          # Typed validated configuration
│   └── storage/           # Desktop atomic storage & Android SAF adapter
├── pkg/
│   └── libgen/            # Reusable core LibGen parser & mirror scraper
├── ui/                    # Fyne native graphical user interface
├── website/               # Tailwind Plus marketing site (Next.js static export)
└── docs/                  # GitHub Pages distribution
```

---

## 🛠️ Building from Source

### Prerequisites
- Go 1.22+
- C compiler (`clang` on macOS, `gcc` on Linux/Windows)
- Fyne CLI (`go install fyne.io/fyne/v2/cmd/fyne@latest`)

### Run in Development
```bash
go run ./cmd/libgen-gui
```

### Build Native Desktop Binary
```bash
go build -o "LibGen Downloader" ./cmd/libgen-gui
```

### Package macOS App Bundle
```bash
fyne package -os darwin -icon icon.png -appID com.libgen.downloader -name "LibGen Downloader"
```

### Cross-Compile for Android (APK)
```bash
fyne-cross android -app-id com.libgen.downloader -icon icon.png
```

---

## 🧪 Testing

```bash
# Run all tests with race detector
go test -race ./...

# Run internal domain tests
go test -v -race ./internal/...

# Run UI & scraper tests
go test -v ./ui ./pkg/libgen
```

---

## 📄 License & Legal Notice
Distributed under the [MIT License](LICENSE).

*Disclaimer: LibGen GUI is an independent search and client tool. Users are solely responsible for verifying the copyright status of files and complying with their local copyright and intellectual property laws.*

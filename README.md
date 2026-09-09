# LibGen Downloader (Fyne GUI)

A lightweight, native cross-platform GUI for discovering, searching, and downloading books and articles from Library Genesis mirrors, built with **Go** and the **Fyne v2** vector toolkit.

---

## 🚀 Features

- **Adaptive Responsive Layout**: Clean UI scaling gracefully across macOS, Windows, Linux, Android (API 35), and iOS.
- **Real-Time Mirror Failover**: Automatic health checks and smart rotation across primary and secondary LibGen/IPFS gateways.
- **Concurrent Download Manager**: Multi-threaded chunked downloads with pause/resume support and live progress bars.
- **Dark / Light Theme Sync**: Follows OS appearance automatically with custom accenting.
- **Zero Heavy Webview**: Compiles to pure native Go binary using OpenGL hardware acceleration.

---

## 📦 Installation & Build

### Prerequisites
- Go 1.22+
- C Compiler (clang on macOS, gcc on Linux/Windows)

### Run in Development
```bash
go run main.go
```

### Build Desktop App (macOS)
```bash
fyne package -os darwin -icon icon.png -appID com.anilpdv.libgendownloader -name "LibGen Downloader"
```

### Cross-Compile for Android (APK)
```bash
fyne-cross android -app-id com.anilpdv.libgendownloader -icon icon.png
```

---

## 🧪 Testing

```bash
# Run unit & widget tests
go test ./... -v

# Run multi-device headless e2e tests
go test -v -run TestDeviceEndToEnd
```

---

## 📄 License
MIT License.

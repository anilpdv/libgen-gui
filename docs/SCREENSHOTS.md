# Screenshot Capture and Optimization Runbook

This document describes how to capture, optimize, and update authentic application screenshots for the LibGen GUI website and repository documentation.

---

## 1. Directory Structure

```
website/src/images/screenshots/
├── desktop-search.webp        # Primary hero & architecture desktop search view
├── mirror-health.webp         # Authentic mirror health & latency prober dialog
└── android-search.webp        # Responsive mobile interface with SAF support

docs/assets/screenshots/       # Static assets mirrored for GitHub Pages and README
├── desktop-search.webp
├── mirror-health.webp
└── android-search.webp
```

---

## 2. Standards & Requirements

Every screenshot displayed on the website or documentation must satisfy:
1. **Authenticity**: Captured directly from running builds of the real Go + Fyne application.
2. **Branding**: Displays official application identifiers (`LibGen Downloader` / `LibGen GUI`).
3. **No Hardware Notch / Fake Frames**: Clean application window with natural OS drop shadows or subtle CSS container borders.
4. **Sanitized Demo Data**: Uses recognizable, academic public domain / open reference titles (e.g. *Designing Data-Intensive Applications*, *Distributed Systems*). No private paths, user tokens, or personal identifiers.
5. **Legibility**: WebP format with quality $\ge 90$, retaining crisp typography and contrast.

---

## 3. Capture Workflow

### macOS Desktop UI
- **Target Application**: LibGen GUI v2.0+
- **Capture Resolution**: $2024 \times 1488$ (Retina @2x) or $1440 \times 900$ (Standard)
- **Steps**:
  1. Launch the application: `go run .`
  2. Perform a search query (e.g., `Distributed Systems`) to populate results with pagination and format badges.
  3. Select a book to activate the download bar.
  4. Capture the window cleanly using macOS window capture (`Cmd + Shift + 4`, then `Space`, then click window).
  5. Save as `screenshot_redesigned.png`.

### Mirror Health Dialog
- **Steps**:
  1. Click the **⚙ Mirrors 2/2** status button in the top search header.
  2. Wait for live latency indicators to display green/healthy status.
  3. Capture the dialog window cleanly.
  4. Save as `screenshot_mirror_dialog.png`.

### Android Mobile View
- **Target Viewport**: $380 \times 720$ (logical) / $1080 \times 2400$ (device)
- **Steps**:
  1. Run with mobile layout or capture from Android emulator running `LibGen-Downloader-Android-arm64.apk`.
  2. Save as `android-search.png`.

---

## 4. Image Optimization and WebP Conversion

Convert raw PNG screenshots to lightweight, lossless/near-lossless WebP images using `cwebp`:

```bash
# 1. Convert desktop search
cwebp -q 90 screenshot_redesigned.png -o website/src/images/screenshots/desktop-search.webp

# 2. Convert mirror health dialog
cwebp -q 90 screenshot_mirror_dialog.png -o website/src/images/screenshots/mirror-health.webp

# 3. Convert Android mobile view
cwebp -q 90 android-search.png -o website/src/images/screenshots/android-search.webp

# 4. Mirror to docs directory
cp website/src/images/screenshots/*.webp docs/assets/screenshots/
```

---

## 5. Website Locations

| Asset | File Path | Component Usage |
|---|---|---|
| `desktop-search.webp` | `website/src/images/screenshots/desktop-search.webp` | `<Hero />`, `<PrimaryFeatures />` |
| `mirror-health.webp` | `website/src/images/screenshots/mirror-health.webp` | `<PrimaryFeatures />` |
| `android-search.webp` | `website/src/images/screenshots/android-search.webp` | `<Hero />` (Mobile preview overlay) |

# tome 📚

`tome` is an interactive, production-ready Terminal User Interface (TUI) PDF bookshelf manager and reader written in Go. It renders PDF covers in a visual grid, supports full mouse interaction, tracks reading progress, and supports terminal graphics protocols (Kitty / ANSI block art fallback).

## ✨ Features

- **Visual Grid View**: Browse your PDF library with rendered cover thumbnails.
- **Rich Metadata Side-Pane**: Displays title, author, page count, and reading progress.
- **Mouse Support**: Single-click to select, scroll wheel to navigate, double-click to read.
- **Integrated Reader Mode**: Full-screen reading with single/dual page view, zoom, and page jumping.
- **Graphics Protocol Support**: Native Kitty Protocol rendering with a high-density ANSI block-art fallback for standard terminals.
- **Concurrent Parsing**: Fast worker pool scans and caches covers using SHA-256 hashing.
- **Search**: Dynamic filtering by title, author, or filename.
- **Open Library Integration**: Fetch missing metadata and covers online.

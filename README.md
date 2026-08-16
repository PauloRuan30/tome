# tome 📚

An interactive **TUI PDF bookshelf manager and reader** for Linux.
Renders your library as a visual cover grid with full mouse support,
tracks reading progress, and reads PDFs *inside the terminal* — with
pixel-perfect pages on Kitty/WezTerm and a readable text/block-art
fallback everywhere else.

## ✨ Features

- **Cover grid** with mouse click/scroll selection and keyboard (hjkl/arrows) navigation
- **Okular mode**: full-page pixel rendering via the Kitty Graphics Protocol
  (out-of-band `/dev/tty` painter, chunked + silent per spec)
- **Readable text mode**: MuPDF text-layer extraction with word-wrap reflow
  for terminals without pixel support (Konsole, etc.)
- **Split-screen metadata pane**: title, author, pages, path, hash,
  progress ("page 23/193 (11%)"), last-opened date
- **Progress tracking** (`~/.config/tome/progress.json`) with resume-on-open
- **Real-time search** (`/`) over title/author/filename
- **Concurrent startup**: SHA-256(64KB) hashing + worker pool + PNG cover cache
- **External open** (`o`) via `xdg-open`

## 🚀 Installation

### Arch Linux (AUR)
```bash
yay -S tome
```

### From source (requires Go 1.22+)
```bash
git clone https://github.com/PauloRuan30/tome && cd tome
make build
sudo make install
```

## 📖 Usage

```bash
tome                          # scan current directory
tome --dir ~/Downloads        # scan a specific directory
tome --clean-cache            # wipe the cover cache
```

## ⌨️ Controls

| Key | Action |
|---|---|
| `hjkl` / arrows | navigate grid / scroll reader |
| mouse wheel | scroll grid / flip or scroll pages |
| click | select book |
| `Enter` / `Space` | open Reader |
| `v` | toggle TEXT ↔ VISUAL reader |
| `d` | dual-page (visual) |
| `:` | jump to page |
| `o` | open in external viewer |
| `/` | search |
| `q` / `Esc` | back / quit |

## 🏗️ Architecture

- **Bubble Tea** (MVU) + **Lipgloss** layout/styling
- **go-fitz** (MuPDF): page rendering, text extraction, metadata
- **Kitty Graphics**: chunked APC payloads written out-of-band to `/dev/tty`
- **ANSI fallback**: 24-bit half-block (`▀`) renderer, 2 pixels per cell

## 🧪 Development

```bash
make test
make build
```

## 📄 License

GPL-3.0-only
```

**2. Add `LICENSE`**
Run this in your terminal to download the standard GPL-3.0 license file:
```bash
curl -o LICENSE https://www.gnu.org/licenses/gpl-3.0.txt
```

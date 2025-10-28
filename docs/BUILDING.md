# Building StratusShell

This document describes the build process for StratusShell.

## Prerequisites

### Required
- **Go 1.24.7 or later** - For building the Go application
- **templ CLI** - For generating Go code from templ templates
  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

### For UI Development
- **Node.js 16+ and npm** - For building the bundled CSS
  - Download from https://nodejs.org/

## Quick Start

The easiest way to build StratusShell is using the Makefile:

```bash
make build
```

This will:
1. Install npm dependencies (if needed)
2. Build the bundled Tailwind CSS + DaisyUI stylesheet
3. Generate Go code from templ templates
4. Compile the Go binary

## Manual Build Steps

If you prefer to build manually or the Makefile doesn't work:

### 1. Build CSS (requires Node.js)

```bash
# Install dependencies
npm install

# Build bundled CSS
npm run build:css
```

This creates `static/bundle.css` containing all Tailwind CSS and DaisyUI styles.

### 2. Generate Templ Files

```bash
templ generate
```

This converts `.templ` files to `*_templ.go` files.

### 3. Build Binary

```bash
go build -o stratusshell main.go
```

## Development

### CSS Development

When developing UI components, you can use the watch mode to automatically rebuild CSS on changes:

```bash
npm run watch:css
```

This watches for changes to:
- `static/input.css`
- `internal/ui/**/*.templ`
- `internal/ui/**/*.go`

### Why Bundled CSS?

We bundle Tailwind CSS and DaisyUI into a single self-hosted file for several reasons:

1. **No CDN Dependency**: The application works in offline and restricted network environments
2. **Faster Loading**: Single CSS file instead of multiple CDN requests
3. **Maintainable**: Clear separation between source CSS (`input.css`) and built output (`bundle.css`)
4. **Smaller Size**: Only includes CSS classes actually used in the application (~59KB minified)
5. **Version Control**: Ensures consistent styling across deployments

### Fallback CSS (Legacy)

The `static/styles.css` file contains a comprehensive fallback stylesheet that was used when the application relied on CDN resources. With the bundled CSS approach, this file is no longer needed for production but is kept for reference.

## Makefile Targets

- `make css` - Build bundled CSS only
- `make generate` - Build CSS and generate templ files
- `make build` - Full build (CSS + templ + Go binary)
- `make test` - Run unit tests
- `make clean` - Remove build artifacts
- `make install` - Install binary and config to system (may require sudo)

## Continuous Integration

For CI/CD pipelines, ensure both Node.js and Go are available:

```yaml
# Example GitHub Actions
steps:
  - name: Setup Node.js
    uses: actions/setup-node@v3
    with:
      node-version: '18'
  
  - name: Setup Go
    uses: actions/setup-go@v4
    with:
      go-version: '1.24.7'
  
  - name: Install templ
    run: go install github.com/a-h/templ/cmd/templ@latest
  
  - name: Build
    run: make build
```

## Troubleshooting

### CSS Not Building

**Error**: `npm: command not found`

**Solution**: Install Node.js from https://nodejs.org/

### Templ Generation Fails

**Error**: `templ: command not found`

**Solution**: Install templ CLI:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

### CSS Classes Not Working

If Tailwind CSS classes aren't being applied:

1. Rebuild the CSS bundle: `npm run build:css`
2. Check that `static/bundle.css` exists and is not empty
3. Ensure the layout template references `/static/bundle.css`
4. Hard refresh your browser (Ctrl+Shift+R or Cmd+Shift+R)

## File Overview

| File | Purpose |
|------|---------|
| `package.json` | npm dependencies and scripts |
| `tailwind.config.js` | Tailwind CSS configuration |
| `static/input.css` | Source CSS with Tailwind directives |
| `static/bundle.css` | Built CSS bundle (committed to repo) |
| `static/styles.css` | Legacy fallback CSS (deprecated) |
| `internal/ui/*.templ` | UI template files |
| `internal/ui/*_templ.go` | Generated Go code (not committed) |

## Distribution

When distributing StratusShell, include:
- The compiled `stratusshell` binary
- The `static/bundle.css` file
- The `configs/default.yaml` file

Users do **not** need Node.js or npm to run StratusShell, only to build it.

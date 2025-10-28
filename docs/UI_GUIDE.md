# StratusShell UI Guide

## Overview

StratusShell features a modern, professional web-based terminal interface built with:
- **Tailwind CSS** for utility-first styling
- **DaisyUI** for consistent UI components
- **HTMX** for dynamic interactions without page reloads

## Key Features

### 🖥️ Multi-Terminal Support
- Create up to **10 concurrent terminal sessions**
- Browser-style tab interface for easy navigation
- Each terminal runs in an isolated GoTTY instance

### 🎨 Modern UI Components

#### Navigation Bar
- **Terminal Menu**: Create new terminals instantly
- **Sessions Menu**: Save and restore terminal layouts
- **Settings Menu**: Configure preferences
- Badge indicator shows max terminal limit

#### Tab Bar
- **Active Tab Highlighting**: Current terminal clearly marked
- **Editable Tab Names**: Click on tab name to rename (blur to save)
- **Close Button**: × button to close individual terminals
- **New Terminal Button**: + button (disabled when at max capacity)
- **Tooltip Support**: Hover over + button for max terminal info

#### Modals
- **Save Session**: Store current terminal configuration
- **Load Session**: Restore previously saved layouts
- **Success/Error Messages**: Clear feedback for all actions

### 🎯 User Interactions

#### Creating Terminals
1. Click **Terminal → New Terminal** in the navbar
2. Or click the **+** button in the tab bar
3. New terminal spawns with auto-generated name (e.g., "Terminal 1")

#### Renaming Terminals
1. Click on the terminal tab name
2. Edit the text in the input field
3. Click outside (blur) or press Enter to save

#### Switching Terminals
- Click on any tab to switch to that terminal
- Active terminal is highlighted with distinct styling

#### Closing Terminals
- Click the × button on any tab
- Confirmation not required (instant close)

#### Saving Sessions
1. Click **Sessions → Save Session...**
2. Enter a session name (required)
3. Add optional description
4. Click **Save**

#### Loading Sessions
1. Click **Sessions → Load Session...**
2. Browse saved sessions with descriptions
3. Click **Load** on desired session
4. Current terminals replaced with saved layout

### 🎨 Design Principles

#### Dark Theme
- Consistent dark mode using DaisyUI's base-200/base-300 colors
- Terminal background: #1e1e1e for authentic terminal feel
- Text optimized for readability

#### Accessibility
- Semantic HTML structure
- ARIA labels where appropriate
- Keyboard navigation support
- High contrast ratios

#### Responsive Design
- Flexbox layouts for fluid resizing
- Horizontal scrolling for many tabs
- Viewport-optimized sizing

## Configuration

### Session Settings
Configure in `configs/default.yaml`:

```yaml
sessions:
  max_terminals: 10              # Maximum concurrent terminals
  auto_reconnect_interval: 10s   # GoTTY reconnect interval
  shell: "bash"                  # Default shell
```

### Theme Customization

The UI uses DaisyUI's theming system. The default theme is "dark", configured in the HTML tag:

```html
<html lang="en" data-theme="dark">
```

Available themes: dark, light, cupcake, cyberpunk, and many more. Change the `data-theme` attribute to switch themes.

### Tailwind Configuration

Custom Tailwind config is embedded in the layout template:

```javascript
tailwind.config = {
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'terminal-bg': '#1e1e1e',
        'terminal-border': '#3e3e42',
        'terminal-text': '#cccccc'
      }
    }
  }
}
```

## Browser Compatibility

Tested and supported on:
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

**Requirements:**
- JavaScript enabled
- WebSocket support
- CSS Grid and Flexbox support

## Performance

- **CDN Delivery**: Tailwind CSS and DaisyUI loaded from CDN for fast initial load
- **Lazy Loading**: Terminal iframes load on demand
- **Efficient Updates**: HTMX only updates changed DOM elements
- **Lightweight**: Custom CSS reduced to minimal overrides (~500 bytes)

## Troubleshooting

### Tabs Not Appearing
- Check browser console for JavaScript errors
- Verify HTMX is loading (check Network tab)
- Ensure `/api/tabs` endpoint is responding

### Styling Issues
- Clear browser cache
- Check if CDN resources are accessible
- Verify Tailwind CSS and DaisyUI are loading

### Terminal Not Loading
- Check GoTTY process is running
- Verify WebSocket connection in browser DevTools
- Check server logs for errors

## Development

### Building
```bash
# Generate Templ files
templ generate

# Build binary
go build -o stratusshell main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run server locally
./stratusshell serve --port=8080
```

### Modifying UI

1. Edit `.templ` files in `internal/ui/`
2. Run `templ generate` to update Go files
3. Rebuild binary
4. Restart server

### Custom Styles

Add custom CSS to `static/styles.css`. This file is loaded after Tailwind and DaisyUI, allowing you to override any default styles.

## Architecture

```
┌─────────────────────────────────────────┐
│         Navbar (DaisyUI)                │
│  Terminal | Sessions | Settings         │
├─────────────────────────────────────────┤
│         Tab Bar (DaisyUI Tabs)          │
│  [Tab 1] [Tab 2] [+]                   │
├─────────────────────────────────────────┤
│                                          │
│        Active Terminal (iframe)         │
│      GoTTY WebSocket Connection         │
│                                          │
│                                          │
└─────────────────────────────────────────┘
```

## Security

- **Authentication**: Session-based auth with HTTP-only cookies
- **CSRF Protection**: All state-modifying requests protected
- **Rate Limiting**: 100 requests/minute per IP
- **Audit Logging**: All terminal operations logged
- **Input Validation**: All user inputs sanitized and validated

## Future Enhancements

Potential improvements for future versions:
- [ ] Drag-and-drop tab reordering
- [ ] Split-pane terminal layouts
- [ ] Custom color themes per terminal
- [ ] Terminal search functionality
- [ ] Command history persistence
- [ ] Collaborative terminal sharing

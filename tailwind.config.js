/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/ui/**/*.templ",
    "./internal/ui/**/*.go",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'terminal-bg': '#1e1e1e',
        'terminal-border': '#3e3e42',
        'terminal-text': '#cccccc'
      }
    },
  },
  plugins: [
    require('daisyui')
  ],
  daisyui: {
    themes: ["dark", "light", "cupcake", "cyberpunk"],
    darkTheme: "dark",
  },
}

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],

  theme: {
    extend: {
      colors: {
        // "Advanced Black" & "Glass Light" Palette via CSS Variables
        background: {
          DEFAULT: 'var(--color-bg-default)',
          subtle: 'var(--color-bg-subtle)',
          card: 'var(--color-bg-card)',
          float: 'var(--color-bg-float)',
        },
        border: {
          DEFAULT: 'var(--color-border-default)',
          subtle: 'var(--color-border-subtle)',
        },
        text: {
          primary: 'var(--color-text-primary)',
          secondary: 'var(--color-text-secondary)',
          tertiary: 'var(--color-text-tertiary)',
        },
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
          DEFAULT: '#0ea5e9', // Sky 500
        },
        accent: {
          DEFAULT: '#6366f1', // Indigo 500
          glow: 'rgba(99, 102, 241, 0.5)',
        }
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      boxShadow: {
        'glow': '0 0 20px rgba(14, 165, 233, 0.15)',
        'glow-strong': '0 0 30px rgba(14, 165, 233, 0.3)',
      },
      backgroundImage: {
        'gradient-dark': 'linear-gradient(to bottom right, #000000, #111111)',
        'gradient-glow': 'var(--gradient-glow)',
      }
    },
  },
  plugins: [],
  // Disable preflight to avoid conflicts with Ant Design
  corePlugins: {
    preflight: false,
  },
}

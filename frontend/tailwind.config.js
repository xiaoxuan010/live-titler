/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#1e293b',
          foreground: '#f8fafc',
        },
        secondary: {
          DEFAULT: '#f1f5f9',
          foreground: '#1e293b',
        },
        destructive: {
          DEFAULT: '#dc2626',
          foreground: '#ffffff',
        },
        border: '#e2e8f0',
        background: '#ffffff',
        foreground: '#0f172a',
      },
    },
  },
  plugins: [],
}

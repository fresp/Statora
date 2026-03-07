/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        operational: '#22c55e',
        degraded: '#eab308',
        partial: '#f97316',
        major: '#ef4444',
        maintenance: '#3b82f6',
      },
    },
  },
  plugins: [],
}

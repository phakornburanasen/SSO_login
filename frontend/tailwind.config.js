/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50:  '#eef5ff',
          100: '#d9e8ff',
          200: '#bcd6ff',
          300: '#8ebcff',
          400: '#5998ff',
          500: '#3478f6',
          600: '#1f5ad8',
          700: '#1b48ad',
          800: '#1c3f8a',
          900: '#1d386f',
        },
      },
    },
  },
  plugins: [],
}

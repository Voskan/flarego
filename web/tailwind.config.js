/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        "flame-red": "#ff6b6b",
        "flame-orange": "#ffa726",
        "flame-yellow": "#ffeb3b",
        "gc-purple": "#b39ddb",
        "heap-teal": "#80cbc4",
        "blocked-salmon": "#ef9a9a",
      },
      fontFamily: {
        mono: ["Monaco", "Menlo", "Ubuntu Mono", "monospace"],
      },
      animation: {
        "pulse-slow": "pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite",
      },
    },
  },
  plugins: [],
};

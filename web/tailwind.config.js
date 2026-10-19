/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        amiguis: {
          yellow: "#FCD116",
          blue: "#003893",
          red: "#CE1126",
          dark: "#121417",
          card: "#1E2229",
          border: "#2A303C",
          muted: "#8F9CAE",
        },
      },
    },
  },
  plugins: [],
};

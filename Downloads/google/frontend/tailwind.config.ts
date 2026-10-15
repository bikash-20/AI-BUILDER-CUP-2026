import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#f0f7ff",
          100: "#e0eefe",
          200: "#b9dcfe",
          300: "#7cc2fc",
          400: "#36a2f8",
          500: "#0d83e8",
          600: "#0266c2",
          700: "#03529c",
          800: "#074783",
          900: "#0c3c6c",
        },
        danger: "#dc2626",
        warning: "#f59e0b",
        success: "#16a34a",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};

export default config;

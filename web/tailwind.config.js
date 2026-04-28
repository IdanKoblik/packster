/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "#0b0b0d",
        surface: "#131318",
        "surface-2": "#17171c",
        border: {
          DEFAULT: "#24242b",
          strong: "#2e2e36",
        },
        ink: {
          DEFAULT: "#e9e9ee",
          dim: "#a4a4ae",
          mute: "#6b6b74",
        },
        accent: {
          DEFAULT: "#d7e27a",
          ink: "#0b0b0d",
        },
        danger: "#ef6a6a",
        ok: "#6bd07a",
      },
      fontFamily: {
        sans: ['"Inter Tight"', "system-ui", "-apple-system", "Segoe UI", "sans-serif"],
        mono: ['"JetBrains Mono"', "ui-monospace", "monospace"],
      },
      letterSpacing: {
        tightish: "-0.005em",
      },
      boxShadow: {
        card: "0 1px 0 rgba(255,255,255,0.03) inset, 0 30px 60px -30px rgba(0,0,0,0.6), 0 8px 20px -12px rgba(0,0,0,0.4)",
      },
      keyframes: {
        fadeIn: {
          from: { opacity: "0", transform: "translateY(4px)" },
          to: { opacity: "1", transform: "none" },
        },
        spin: { to: { transform: "rotate(360deg)" } },
        shimmer: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(100%)" },
        },
      },
      animation: {
        "fade-in": "fadeIn .25s ease",
        spin: "spin .7s linear infinite",
        shimmer: "shimmer 1.6s infinite",
      },
    },
  },
  plugins: [],
};

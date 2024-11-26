/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/html/**/*.html"],
  theme: {
    extend: {
      maxWidth: {
        edges: "95rem",
      },
    },
  },
  plugins: [require("@tailwindcss/forms")],
};

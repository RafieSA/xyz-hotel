/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#8B5A2B',
        'primary-hover': '#6F4620',
        'primary-light': '#FDF6EC',
        cream: '#FDF6EC',
        teal: '#1A3A4A',
        gold: '#C9A86A',
      },
      borderRadius: {
        xl: '12px',
        '2xl': '16px',
      },
      fontFamily: {
        display: ['Playfair Display', 'Poppins', 'serif'],
        body: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}

export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    darkMode: 'class',
    theme: {
        extend: {
            colors: {
                page: {
                    DEFAULT: '#000000',
                    subtle: '#0a0a0a',
                    card: '#111111',
                    float: '#1a1a1a',
                },
                border: {
                    DEFAULT: '#262626',
                    subtle: '#1f1f1f',
                },
                text: {
                    primary: '#ffffff',
                    secondary: '#a1a1aa',
                    tertiary: '#52525b',
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
                    DEFAULT: '#0ea5e9',
                },
            },
            fontFamily: {
                sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
            },
        },
    },
    plugins: [],
}

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["templates/**/*.{html,js,templ}",],
    theme: {
        extend: {
            clifford: "#da373d",
            // Note: browser shrinks font-family: monospace, so using it 2 times avoids it.
            fontFamily: {
                sans: ['"JetBrains Mono"', "monospace", "monospace"],
            },
            colors: {
                terminal: {
                    50: "#f8fdf4",
                    100: "#e4f8d4",
                    200: "#c8eaa9",
                    300: "#a3d981",
                    400: "#7cbf5b",
                    500: "#5aa642",
                    600: "#468b31",
                    700: "#376f28",
                    800: "#2e5b23",
                    900: "#25481e",
                    950: "#113115",
                },
            },
        },
    },
    plugins: [
        /**
         * Custom variant function for Tailwind CSS.
         *
         * @param {object} options - Destructured object with the `addVariant` function.
         * @param {Function} options.addVariant - The `addVariant` function from Tailwind CSS.
         *
         * @see {@link https://github.com/MrChip53/blog.simoni.dev/blob/2803be5001c78149df01aa9c2ada8d5c0b21a747/tailwind.config.js#L30}
         */
        function ({ addVariant }) {
            addVariant("child", "& > *");
        },
    ],
};

package web

import "embed"

//go:generate tailwindcss -c tailwind.config.js -i style.css -o embed/assets/css/style.css --minify
//go:embed embed
var Static embed.FS

cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./html
GOOS=js GOARCH=wasm go build -o=html/game.wasm .
zip -r html.zip ./html

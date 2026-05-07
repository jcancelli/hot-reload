(() => {
	const url = new URL("{{.WebSocketRoute}}", "ws://localhost:{{.Port}}")
	const ws = new WebSocket(url)
	ws.onmessage = () => {
		window.location.reload()
	}
})()

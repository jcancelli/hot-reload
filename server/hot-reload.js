(() => {
	const url = new URL("{{.WebSocketRoute}}", "ws://localhost:{{.Port}}")
	const ws = new WebSocket(url)

	let lastPingTimeout = null
	function ping() {
		ws.send(JSON.stringify({
			kind: "ping",
		}))
		lastPingTimeout = setTimeout(() => {
			alert("Dev server disconnected")
		}, 5_000)
	}

	setInterval(ping, 10_000)

	ws.onmessage = (event) => {
		const data = event.data
		switch (data.kind) {
			case "reload":
				window.location.reload()
				break
			case "pong":
				clearTimeout(lastPingTimeout)
				lastPingTimeout = null
				break
			default:
				console.log(event)
				break
		}
	}
})()

(() => {
	const url = new URL("{{.WebSocketRoute}}", "ws://localhost:{{.Port}}")

	/** @type {WebSocket} */
	let ws

	function connect() {
		ws = new WebSocket(url)

		ws.onmessage = (event) => {
			if (event.data === "RELOAD") {
				window.location.reload()
			} else {
				alert(`Unexpected message ${event}`)
			}
		}

		ws.onopen = () => {
			console.log("[hot-reload] connected")
		}

		ws.onclose = () => {
			console.log("[hot-reload] disconnected")
		}

		ws.onerror = (event) => {
			console.log(`[hot-reload] websocket error: ${event.toString()}`)
			ws.close()
			console.log(`[hot-reload] trying to reconnect`)
			connect()
		}
	}

	connect()
})()

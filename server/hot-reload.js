(() => {
	const url = new URL("{{.WebSocketRoute}}", "ws://localhost:{{.Port}}")

	const MAX_RECONNECT_RETRIES = 5
	let reconnectRetries = 0

	/** @type {WebSocket} */
	let ws

	function connect() {
		if (reconnectRetries > MAX_RECONNECT_RETRIES) {
			console.log("[hot-reload] failed to reconnect (max retries reached)")
		}

		ws = new WebSocket(url)

		ws.onmessage = (event) => {
			if (event.data === "RELOAD") {
				ws.close()
				window.location.reload()
			} else {
				alert(`Unexpected message ${event}`)
			}
		}

		ws.onopen = () => {
			console.log("[hot-reload] connected")
			reconnectRetries = 0
		}

		ws.onclose = () => {
			console.log("[hot-reload] disconnected")
		}

		ws.onerror = (event) => {
			console.log(`[hot-reload] websocket error: ${event.toString()}`)
			ws.close()
			console.log(`[hot-reload] trying to reconnect`)
			reconnectRetries += 1
			connect()
		}
	}

	connect()
})()

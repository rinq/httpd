window.onload = function () {
    ws = new WebSocket("ws://localhost:8081", ["rinq-1.0+cbor", "rinq-1.0+json"])

    ws.onopen = function () {
        console.log("open", ws.protocol)
    }
    ws.onmessage = function () {
        console.log("message")
    }
    ws.onclose = function () {
        console.log("close")
    }

    console.log("ws")
}

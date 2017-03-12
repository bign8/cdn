var ws = new WebSocket("ws://" + window.location.host + "/ws");

ws.onopen = function() {
  ws.send(JSON.stringify({msg: "hello server!"}))
}

ws.onmessage = function(event) {
  var m = JSON.parse(event.data);
  console.debug("Received message", m.msg);
}

ws.onerror = function(event) {
  console.debug(event)
}

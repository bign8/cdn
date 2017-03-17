var ws, ticker, framer, last;

function start() {
  ws = new WebSocket("ws://" + window.location.host + "/ws/admin");
  ws.onopen = function() {
    last = 1;
    ping();
  }

  ws.onmessage = function(event) {
    var m = JSON.parse(event.data);
    console.debug("Received message", m.typ);
  }

  ws.onerror = function(event) { console.debug(event); };

  ws.onclose = function(event) {
    console.debug("closing", event);
    window.cancelAnimationFrame(framer);
    window.clearTimeout(ticker);
    last *= 2;
    window.setTimeout(start, Math.floor(Math.random() * last));
  }
}

function pong() {
  ws.send(JSON.stringify({typ: "ping"}));
  ticker = window.setTimeout(ping, 30 * 1000);
}

function ping() {
  framer = window.requestAnimationFrame(pong);
}

start();

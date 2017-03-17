// Chart documentation: http://www.chartjs.org/docs/

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
    myChart.data.datasets[0].data.push({
      x: Date.now(),
      y: 99,
    });
    myChart.update();
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

var data = {
    datasets: [
        {
            label: "Type 1 (99)",
            fill: false,
            borderColor: "rgba(75,192,192,1)",
            data: [{
              x: Date.now() - 2 * 1000,
              y: 99,
            }, {
              x: Date.now()- 1000,
              y: 80,
            }, {
              x: Date.now(),
              y: 81,
            }],
        }, {
          label: "Type 2 (23)",
          fill: false,
          data: [{
            x: Date.now(),
            y: 62,
          }, {
            x: Date.now() + 1000,
            y: 68,
          }, {
            x: Date.now() + 2 * 1000,
            y: 99,
          }],
        }
    ]
};

var options = {
  responsive: false, // don't resize
  scales: {
    xAxes: [{
      type: "time",
      time: {
        unit: 'second',
      },
    }]
  },
  tooltips: {
    enabled: false,
  //   callbacks: {
  //     title: function(items, data) {
  //       // TODO: convert xlabel to pretty string
  //       var obj = items[0];
  //       return data.datasets[obj.datasetIndex].data[obj.index].x;
  //     },
  //   }
  }
};

var chart = document.getElementById('chart');
var myChart = new Chart(chart, {
  type: 'line',
  data: data,
  options: options
});

// TODO: make all this smart

function checkSetIdx(idx, now) {
  var obj = myChart.data.datasets[idx];
  if (now - obj.data[obj.data.length - 1].x > 1500) {
    obj.data.push({
      x: now,
      y: 0,
    });
  }
  if (obj.data.length > 10) {
    obj.data.shift();
  }
}

function setZeros() {
  console.log("setting zeros");
  var now = Date.now();
  checkSetIdx(0, now);
  checkSetIdx(1, now);
  myChart.update();
  window.setTimeout(window.requestAnimationFrame.bind(this, setZeros), 1000);
}

setZeros();

// Chart documentation: http://www.chartjs.org/docs/

var ws, ticker, framer, last, brains;

function start() {
  ws = new WebSocket("ws://" + window.location.host + "/ws/admin");
  ws.onopen = function() {
    last = 1;
    ping();
  }

  ws.onmessage = function(event) {
    var m = JSON.parse(event.data);
    switch (m.typ) {
      case 'stat': brain.addStat(m); break;
      case 'pong': console.debug("Received PONG"); break;
      default: console.debug("Received unknown message type", m.typ); break;
    }
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

function createChart(ctx, name) {
  return new Chart(ctx, {
    type: 'line',
    data: {
      datasets: [],
    },
    options: {
      responsive: false, // don't resize
      title: {
        display: true,
        text: name,
      },
      scales: {
        xAxes: [{
          type: "time",
          display: false,
        }],
        yAxes: [{
          ticks: {
            beginAtZero: true,
          },
          // scaleLabel: {
          //   display: true,
          //   labelString: "count",
          // },
        }],
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
    }
  })
}

function createDataset(name, color) {
  return {
    label: name,
    fill: false,
    borderColor: color,
    data: [],
  };
}

var myChart = createChart(document.getElementById('chart'), 'Origin xxxx');
myChart.data.datasets.push(createDataset('Type 1', 'rgba(75,192,192,1)'));
myChart.data.datasets.push(createDataset('Type 2', ''));

// Setup mock data
var now = Date.now();
myChart.data.datasets[0].data = [{x: now - 2 * 1000, y: 99}, {x: now - 1000, y: 80}, {x: now, y: 81}];
myChart.data.datasets[1].data = [{x: now, y: 62}, {x: now + 1000, y: 68}, {x: now + 2 * 1000, y: 99}];
myChart.update();

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

function Brain() {
  this.kinds = ['origin', 'server', 'client']; // TODO: allow re-ordering
  this.charts = {}; // map of ids to chart types
  this.stats = {}; // map of ids to stat metadata
}

Brain.prototype.addStat = function(stat) {
  // {"typ":"stat","msg":{"img":{"tot":4,"new":1}},"who":{"id":"C417A89F4D8B2FA392595277A31A41B3","kind":"origin"}}
  console.log(stat);

  // Initalize chart object if necessary
  if (!this.charts.hasOwnProperty(stat.who.id)) {
    console.log('TODO: initalize chart');
    var ctx = document.createElement('canvas');
    ctx.width = 800;
    ctx.height = 400;
    document.body.appendChild(ctx); // TODO: figure out the correct place to put this chart
    this.charts[stat.who.id] = createChart(ctx, stat.who.kind + " " + stat.who.id.substr(0, 10).toLowerCase());
  } else {
    console.log('chart initailized');
  }
  myChart.data.datasets[0].data.push({
    x: Date.now(),
    y: 99,
  });
  myChart.update();
};

// Initalize the brain
brain = new Brain();

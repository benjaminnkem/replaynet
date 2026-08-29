const timeline = document.getElementById('timeline');
const source = new EventSource('/events');

source.onmessage = function(e) {
  const msg = JSON.parse(e.data);
  const row = document.createElement('div');
  row.className = 'row ' + msg.type;

  if (msg.type === 'request') {
    row.textContent = msg.offsetMs + 'ms  ' + msg.method + ' ' + msg.path;
  } else {
    const cls = msg.status >= 500 ? 'status-error' : (msg.status >= 400 ? 'status-warn' : 'status-ok');
    row.innerHTML = msg.offsetMs + 'ms  <span class="' + cls + '">&larr; ' + msg.status + '</span>';
  }

  timeline.appendChild(row);
  timeline.scrollTop = timeline.scrollHeight;
};

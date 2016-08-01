window.onload = function() {
  var conn;
  var msg = document.getElementById("msg");
  var log = document.getElementById("log");

  function appendLog(item) {
    var doScroll = log.scrollTop === log.scrollHeight - log.clientHeight;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }
  document.getElementById("form").onsubmit = function() {
    if (!conn) {
      return false;
    }
    if (!msg.value) {
      return false;
    }
    conn.send(msg.value);
    msg.value = "";
    return false;
  };
  if (window["WebSocket"]) {
    new_uri = "ws://" + window.location.host + "/ws";
    conn = new WebSocket(new_uri)
    conn.onclose = function(evt) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Connection closed.</b>";
      appendLog(item);
    };
    conn.onmessage = function(evt) {
      var messages = evt.data.split('\n');
      for (var i = 0; i < messages.length; i++) {
        var item = document.createElement("div");
        item.innerText = messages[i];
        appendLog(item);
      }
    };
  } else {
    var item = document.createElement("div");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    appendLog(item);
  }
  document.getElementById("makeadmin").onclick = function() {
    conn.send("ADMIN foobar");
  }
  document.getElementById("getstatus").onclick = function() {
    conn.send("STATUS");
  }
  document.getElementById("playvideo").onclick = function() {
    conn.send("NEW BigBuckBunny 1234");
  }
  document.getElementById("setposition").onclick = function() {
    conn.send("POS " + Math.floor((Math.random() * 1000) + 1));
  }
};

// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func wsEcho(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{} // use default options

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func wsHTML(w http.ResponseWriter, _ *http.Request) {
	const html = `<!DOCTYPE html>
<html>
<body>
<h1>You should get a new message every second</h1>
<script type="text/javascript">
    const ws = new WebSocket('ws://' + location.host + '/ws/echo');
    ws.onopen = function(e) {
        document.body.innerHTML += 'WS opened<br>';
    };
    ws.onmessage = function(e) {
        document.body.innerHTML += e.data + '<br>';
    };
    function send() {
        ws.send(document.getElementById('message').value);
    }
</script>
<button onclick="ws.close()">Close Source</button>
<br/>
<input type="text" id="message" value="" /><button onclick="send()">Send</button>
<br/>
</body>
</html>
`
	w.Write([]byte(html))
}

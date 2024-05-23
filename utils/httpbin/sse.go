// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func events(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/events/"):]

	ms, ok := atoi(w, p)
	if !ok {
		return
	}

	// Make sure that the writer supports flushing.
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Transfer-Encoding", "identity")

	var i int
	for {
		if err := r.Context().Err(); err != nil {
			log.Printf("client disconnected (%s)", err)
			return
		}

		i++
		fmt.Fprintf(w, "data: Message: %d\n\n", i)
		f.Flush()
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func eventsHTML(w http.ResponseWriter, _ *http.Request) {
	const html = `<!DOCTYPE html>
<html>
<body>
<h1>You should get a new message every second</h1>
<script type="text/javascript">
    var source = new EventSource('/events/1000');
    source.onmessage = function(e) {
        document.body.innerHTML += e.data + '<br>';
    };
</script>
<button onclick="source.close()">Close Source</button>
<br>
</body>
</html>
`
	w.Write([]byte(html))
}

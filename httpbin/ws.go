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

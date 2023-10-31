package main

import (
	"github.com/gorilla/websocket"
	"log"
	"log-transfer/util"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")
	defer ws.Close()

	req := Request{}

	err = ws.ReadJSON(&req)
	if err != nil {
		log.Println(err)
	}

	if req.Namespace == "" || req.Pod == "" {
		log.Println("wrong request lack of namespace or pod name")
		return
	}

	podLog := util.GetPodLog(req.Pod, req.Namespace)
	defer podLog.Close()

	for {
		buf := make([]byte, 1024)
		num, err := podLog.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		err = ws.WriteMessage(websocket.TextMessage, buf[:num])
		if err != nil {
			log.Println(err)
			return
		}
	}

}

func main() {
	http.HandleFunc("/ws", wsEndpoint)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Request struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}

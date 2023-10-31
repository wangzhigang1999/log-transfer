package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"log-transfer/util"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		log.Println(r.Host)
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

	// tail line must be positive and less than 100
	if req.TailLines <= 0 || req.TailLines > 100 {
		req.TailLines = 100
	}

	stream := util.GetPodLog(req.Pod, req.Namespace, &req.TailLines)
	defer stream.Close()

	for {
		buf := make([]byte, 2048)
		num, err := stream.Read(buf)

		if err == io.EOF {
			log.Println("stream end")
			break
		}

		if num == 0 {
			log.Println("no data")
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			log.Println(err)
			break
		}
		err = ws.WriteMessage(websocket.TextMessage, buf[:num])
		if err != nil {
			log.Println(err)
			break
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
	TailLines int64  `json:"tailLines"`
}

package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"log-transfer/util"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var allowNamespace = sync.Map{}

func init() {
	allowNamespace.Store("schedule", true)
	allowNamespace.Store("train-job", true)
	allowNamespace.Store("wanz", true)
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
		ws.WriteMessage(websocket.TextMessage, []byte("wrong request"))
		return
	}

	if req.Namespace == "" || req.Workload == "" {
		log.Println("wrong request lack of namespace or pod name")
		ws.WriteMessage(websocket.TextMessage, []byte("wrong request lack of namespace or pod name"))
		return
	}

	// check namespace
	if _, ok := allowNamespace.Load(req.Namespace); !ok {
		log.Println("namespace not allowed")
		ws.WriteMessage(websocket.TextMessage, []byte("namespace not allowed"))
		return
	}

	// tail line must be positive and less than 100
	if req.TailLines <= 0 || req.TailLines > 100 {
		req.TailLines = 100
	}
	var stream io.ReadCloser
	if req.Mode == "job" {
		stream, err = util.GetJobLog(req.Workload, req.Namespace, &req.TailLines)
	} else {
		stream, err = util.GetPodLog(req.Workload, req.Namespace, &req.TailLines)
	}

	defer stream.Close()

	if err != nil {
		log.Println(err)
		ws.WriteMessage(websocket.TextMessage, []byte("get log error,error:"+err.Error()))
		return
	}

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
	Workload  string `json:"workload"`
	TailLines int64  `json:"tailLines"`
	Mode      string `json:"mode"`
}

package pkg

import (
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func TransferLog(ws *websocket.Conn, req *TargetWorkloadMSG) {

	if req.TailLines <= 0 || req.TailLines > MaxTailLines {
		req.TailLines = MaxTailLines
	}

	var stream io.ReadCloser
	var err error

	if req.Mode == ModeJob {
		stream, err = GetJobLog(req.Workload, req.Namespace, &req.TailLines)
	} else if req.Mode == ModePod {
		stream, err = GetPodLog(req.Workload, req.Namespace, &req.TailLines)
	}

	if err != nil {
		slog.Error("get log error", "reason", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte("get log error,error:"+err.Error()))
		return
	}
	defer stream.Close()

	for ws != nil {
		buf := make([]byte, 2048)
		num, err := stream.Read(buf)

		if err == io.EOF {
			_ = ws.WriteMessage(websocket.TextMessage, []byte("\nNo more content to read. Typically this means the container has crashed or the job has finished running."))
			return
		}

		if err != nil && err != io.EOF {
			slog.Error("read log error", "reason", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte("read log error,error:"+err.Error()))
			return
		}

		if num == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		_ = ws.WriteMessage(websocket.TextMessage, buf[:num])
	}
}

func HandleRequest(ws *websocket.Conn, msg *TargetWorkloadMSG) error {
	errorCount := 0
	for {
		// do not read message from client if error count is greater than MaxErrorCount
		if errorCount >= MaxErrorCount {
			return errors.New("error count is greater than MaxErrorCount")
		}

		err := ws.ReadJSON(&msg)
		if err != nil {
			slog.Warn("wrong request", "reason", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte("wrong request,error:"+err.Error()))
			errorCount++
			continue
		}

		if valid, err := msg.Valid(); !valid {
			slog.Warn("wrong request", "reason", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			errorCount++
			continue
		}

		slog.Info("accept request", "request", msg)
		break
	}
	go HandleCMD(ws)
	return nil
}

func HandleCMD(ws *websocket.Conn) {
	for {
		messageType, bytes, err := ws.ReadMessage()
		if err != nil {
			slog.Error("read messageType error,will close connection", "error", err)
			ws.Close()
			break
		}
		slog.Info("receive message", "messageType", messageType, "message", string(bytes))
	}
}

func OriginChecker(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	slog.Info("check origin", "origin", origin)
	return true
}

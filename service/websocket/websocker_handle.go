package websocket

import (
	"fmt"
	"net/http"

	"github.com/yumimobi/trace/util/json"

	"github.com/gorilla/websocket"
	"github.com/yumimobi/trace/config"
	"github.com/yumimobi/trace/log"
	"github.com/yumimobi/trace/service/grpc"
)

var upgrader = websocket.Upgrader{
	// 跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func WebSocketInit() {
	conf := config.Conf

	http.HandleFunc("/trace", WebSocketHandler)
	fmt.Println("websocket start ", conf.Server.WebSocket.Address+":"+conf.Server.WebSocket.Port)

	err := http.ListenAndServe(conf.Server.WebSocket.Address+":"+conf.Server.WebSocket.Port, nil)
	if err != nil {
		log.Entry.Error().Err(err).Msg("websocket listen and serve is failed")
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Entry.Error().Err(err).Msg("websocket connect is failed")
	}
	defer c.Close()

	log.Entry.Debug().Msg("websocket is start.")

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Entry.Error().Err(err).Int("msg type", mt).Msg("websocket read is failed")
			break
		}
		fmt.Println("websocket read msg is:", string(message))

		log.Entry.Debug().Str("req", string(message)).Msg("websocket request msg")

		resp, err := convertMsgFormat(message)
		if err != nil {
			log.Entry.Error().Err(err).Str("req", string(message)).Msg("websocket convert msg is failed")
			break
		}

		log.Entry.Debug().Str("resp", string(resp)).Msg("websocket response msg")

		err = c.WriteMessage(mt, resp)
		if err != nil {
			log.Entry.Error().Err(err).Int("msg type", mt).Msg("websocket write is failed")
			break
		}
	}
}

func convertMsgFormat(req []byte) ([]byte, error) {
	request := &grpc.Request{}
	err := json.Unmarshal(req, request)
	if err != nil {
		return nil, err
	}

	response, err := grpc.SendMsg(request)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

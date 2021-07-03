package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tech-With-Tim/Socket-Api/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

var s *Server
var testServer *httptest.Server

func TestMain(m *testing.M) {
	s = CreateServer()
	var err error
	s.Config, err = utils.LoadConfig("../", "test")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = s.RegisterRedisHandler("challenges.new", NewChallengeSub(s))
	if err != nil {
		log.Fatal(err)
	}

	err = s.RegisterCommand("1", func(sender *Client, request Request) error {
		msg, err := json.Marshal(request.Data)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		s.RedisClient.Publish(context.Background(), "challenges.new", msg)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	go RedisHandler(s)

	//http.HandleFunc("/", HandleConnections(s))
	testServer = httptest.NewServer(http.HandlerFunc(HandleConnections(s)))
	defer testServer.Close()

	os.Exit(m.Run())
}

func pingHandler(t *testing.T, wg *sync.WaitGroup){
	defer wg.Done()
	url := "ws" + strings.TrimPrefix(testServer.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	for i:=0; i<5; i++ {
		if err := ws.WriteJSON(map[string]interface{}{"op":"0"}); err != nil {
			require.NoError(t, err)
		}
		expectedRes := &pingResponse{Operation: "0"}
		res := &pingResponse{}
		err := ws.ReadJSON(res)
		require.NoError(t, err)
		require.Equal(t, expectedRes, res)
	}
}

func publishRedis(t *testing.T, wg *sync.WaitGroup){
	defer wg.Done()
	url := "ws" + strings.TrimPrefix(testServer.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()
	for i:=0; i<5; i++ {
		err = ws.WriteJSON(map[string]interface{}{"op": "1",
			"d": map[string]interface{}{
				"hello": "world",
			},
		})
		require.NoError(t, err)
		expectedRes := redisReponse{Response: "world"}
		//_, p ,_ := ws.ReadMessage()
		var res redisReponse
		err = ws.ReadJSON(&res)
		require.NoError(t, err)
		if res.Response != ""{
			require.Equal(t, expectedRes, res)
		}
		// fmt.Println(string(p))

	}

}
func TestSockets(t *testing.T) {
	var wg sync.WaitGroup
	//Test pings
	for i:=0; i<5; i++{
		wg.Add(1)
		go pingHandler(t, &wg)
		fmt.Println("HI")
	}
	wg.Wait()
	wg.Add(1)
	go publishRedis(t, &wg)

	wg.Wait()
}


type pingResponse struct {
	Operation string `json:"op"`
}

type redisReponse struct {
	Response  string `json:"hello"`
}
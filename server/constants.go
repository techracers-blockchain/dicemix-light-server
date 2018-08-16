package server

import (
	"sync"
	"time"

	"../commons"
	"github.com/gorilla/websocket"
)

const (
	// rounds
	joinTXRound = 1

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// number of peers required to start DiceMix protocol
	maxPeers = 3

	// Time to wait for response from peers.
	responseWait = 5
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients    map[*Client]int32
	peers      []*commons.PeersInfo
	request    chan []byte
	register   chan *Client
	unregister chan *Client
	nextState  int
	sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

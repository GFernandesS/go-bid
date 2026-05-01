package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/GFernandesS/go-bid/internal/usecase/bids"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	//Request
	PlaceBid MessageKind = iota

	//Success
	SuccessfullyPlaceBid

	//Errors
	FailedToPlaceBid
	InvalidJSON

	//Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Content string      `json:"message,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserId  uuid.UUID   `json:"user_id,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	Clients    map[uuid.UUID]*Client

	BidsService *BidsService
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsService *BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		Context:     ctx,
		BidsService: bidsService,
	}
}

const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	pingPeriod     = (readDeadline * 9) / 10 //90% from readDeadLine
	writeDeadline  = 10 * time.Second
)

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	UserId uuid.UUID
	Send   chan Message
}

func (client *Client) ReadEventLoop() {
	defer func() {
		client.Room.Unregister <- client
		_ = client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)

	_ = client.Conn.SetReadDeadline(time.Now().Add(readDeadline))

	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		var m Message

		m.UserId = client.UserId

		err := client.Conn.ReadJSON(&m)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("unexpected close error on websocket", "error", err)

				return
			}

			client.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Content: "Invalid JSON",
				UserId:  m.UserId,
			}

			continue
		}

		client.Room.Broadcast <- m
	}
}

func (client *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()

		_ = client.Conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeDeadline))

			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("failed to write ping", "error", err)

				return
			}

		case message, ok := <-client.Send:
			if !ok {
				_ = client.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Content: "closing websocket connection",
				})

				return
			}

			if message.Kind == AuctionFinished {
				close(client.Send)
				return
			}

			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeDeadline))

			err := client.Conn.WriteJSON(message)

			if err != nil {
				client.Room.Unregister <- client
				return
			}
		}

	}
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		UserId: userId,
		Send:   make(chan Message, 512),
	}
}

func (auctionLobby *AuctionLobby) AddRoom(room *AuctionRoom) {
	auctionLobby.Lock()

	defer auctionLobby.Unlock()

	auctionLobby.Rooms[room.Id] = room
}

func (room *AuctionRoom) registerClient(client *Client) {
	slog.Info("new user connected", "Client", client)

	room.Clients[client.UserId] = client
}

func (room *AuctionRoom) unregisterClient(client *Client) {
	slog.Info("user disconnected", "Client", client)

	delete(room.Clients, client.UserId)
}

func (room *AuctionRoom) broadcastMessage(message Message) {
	slog.Info("new message received", "RoomId", room.Id, "Message", message.Content, "UserId", message.UserId)

	switch message.Kind {
	case PlaceBid:
		bid, err := room.BidsService.PlaceBid(room.Context, bids.PlaceBidRequest{
			ProductId: room.Id,
			Amount:    message.Amount,
		}, message.UserId)

		if err != nil {
			if errors.Is(err, InsufficientBidError) {
				if client, ok := room.Clients[message.UserId]; ok {
					client.Send <- Message{
						Content: InsufficientBidError.Error(),
						Kind:    FailedToPlaceBid,
					}
				}

				return
			}
		}

		if client, ok := room.Clients[message.UserId]; ok {
			client.Send <- Message{
				Content: "Your bid was successfully placed",
				Kind:    SuccessfullyPlaceBid,
				UserId:  message.UserId,
			}
		}

		for id, client := range room.Clients {
			if id == message.UserId {
				continue
			}

			newBidMessage := Message{
				Content: "A new bid was placed",
				Kind:    NewBidPlaced,
				Amount:  bid.BidAmount,
				UserId:  message.UserId,
			}

			client.Send <- newBidMessage
		}
	case InvalidJSON:
		client, ok := room.Clients[message.UserId]

		if !ok {
			slog.Info("client not found", "UserId", message.UserId)
			return
		}

		client.Send <- message
	}
}

func (room *AuctionRoom) Run() {
	slog.Info("Auction has begun", "auctionId", room.Id)

	defer func() {
		close(room.Broadcast)
		close(room.Register)
		close(room.Unregister)
	}()

	for {
		select {
		case client := <-room.Register:
			room.registerClient(client)

		case client := <-room.Unregister:
			room.unregisterClient(client)

		case message := <-room.Broadcast:
			room.broadcastMessage(message)

		case <-room.Context.Done():

			slog.Info("auction has ended", "AuctionId", room.Id)

			for _, client := range room.Clients {
				client.Send <- Message{
					Content: "The auction has ended",
					Kind:    AuctionFinished,
				}
			}

			return
		}
	}
}

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/stojg/cyberspace/lib/client"
	"github.com/stojg/cyberspace/lib/core"
	"github.com/stojg/vector"
	"golang.org/x/net/websocket"
)

type clientHub struct {
	sync.Mutex
	clients []*client.Client
}

func (network *clientHub) remove(i int) {
	network.clients = append(network.clients[:i], network.clients[i+1:]...)
}

func (network *clientHub) add(c *client.Client) {
	network.Lock()
	network.clients = append(network.clients, c)
	network.Unlock()
}

func (network *clientHub) Write(cmd client.MessageType, data []byte) (int, error) {
	network.Lock()
	// bytes written
	n := 0
	defer network.Unlock()
	for i, c := range network.clients {
		nc, err := c.Update(data, cmd)
		n += nc
		if err != nil {
			Printf("network error, removing client\n")
			c.Close()
			network.remove(i)
			return n, err
		}
	}
	return n, nil
}

func initNetwork(lvl *level) *clientHub {
	Println("Inititalising Network")

	hub := &clientHub{
		clients: make([]*client.Client, 0),
	}
	ch := client.NewClientManager()
	http.Handle("/ws/", websocket.Handler(ch.Websocket))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		if r.URL.Path[1:] == "" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		http.ServeFile(w, r, "static/"+r.URL.Path[1:])
	})

	go func(client chan *client.Client, h *clientHub) {
		for newClient := range client {
			Println("client connected")
			_, err := newClient.Update(lvl.fullDraw().Bytes(), 1)
			if err != nil {
				Println("network error, ignoring new client..")
			} else {
				h.add(newClient)
			}

		}
	}(ch.NewClients(), hub)

	go func() {
		Println(http.ListenAndServe(":8080", nil))
	}()

	return hub

}

const (
	_ byte = iota
	instEntityID
	instPosition
	instOrientation
	instType
	instScale
)

func (l *level) draw() *bytes.Buffer {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, float32(atomic.LoadUint64(&currentFrame)))
	if err != nil {
		Printf("draw() error %s", err)
	}

	for _, graphic := range core.List.Graphics() {
		gameObject := graphic.GameObject()
		if !graphic.IsRendered() {
			serialize(buf, gameObject)
			graphic.SetRendered()
			continue
		}

		body := gameObject.Body()
		if body != nil && body.Awake() {
			serialize(buf, gameObject)
		}
	}
	return buf
}

func (l *level) fullDraw() *bytes.Buffer {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, float32(atomic.LoadUint64(&currentFrame)))
	if err != nil {
		Printf("fullDraw() error %s", err)
	}
	for _, graphic := range core.List.Graphics() {
		gameObject := graphic.GameObject()
		serialize(buf, gameObject)
	}
	return buf
}

func (l *level) drawDead() *bytes.Buffer {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, float32(atomic.LoadUint64(&currentFrame)))
	if err != nil {
		Printf("drawDead() error %s", err)
	}

	for _, id := range core.List.Deleted() {
		if err := binaryStream(buf, instEntityID, id); err != nil {
			Printf("binarystream error %s", err)
		}
	}
	core.List.ClearDeleted()

	return buf
}

func serialize(buf io.Writer, gameObject *core.GameObject) {
	if err := binaryStream(buf, instEntityID, gameObject.ID()); err != nil {
		Printf("binarystream error %s", err)
	}
	if err := binaryStream(buf, instPosition, gameObject.Transform().Position()); err != nil {
		Printf("binarystream error %s", err)
	}
	if err := binaryStream(buf, instOrientation, gameObject.Transform().Orientation()); err != nil {
		Printf("binarystream error %s", err)
	}

	graphic := gameObject.Graphic()
	if graphic != nil {
		if err := binaryStream(buf, instType, graphic.Model()); err != nil {
			Printf("binarystream error %s", err)
		}
	}
	if err := binaryStream(buf, instScale, gameObject.Transform().Scale()); err != nil {
		Printf("binarystream error %s", err)
	}
}

func binaryStream(buf io.Writer, varType byte, value interface{}) error {
	var err error
	if err = binary.Write(buf, binary.LittleEndian, varType); err != nil {
		return err
	}
	switch val := value.(type) {
	case uint8:
		err = binary.Write(buf, binary.LittleEndian, val)
	case uint16:
		err = binary.Write(buf, binary.LittleEndian, float32(val))
	case uint32:
		err = binary.Write(buf, binary.LittleEndian, float32(val))
	case int:
		err = binary.Write(buf, binary.LittleEndian, int32(val))
	case float32:
		err = binary.Write(buf, binary.LittleEndian, val)
	case float64:
		err = binary.Write(buf, binary.LittleEndian, float32(val))
	case core.ID:
		err = binary.Write(buf, binary.LittleEndian, float32(val))
	case *vector.Vector3:
		if err = binary.Write(buf, binary.LittleEndian, float32(val[0])); err != nil {
			return err
		}
		if err = binary.Write(buf, binary.LittleEndian, float32(val[1])); err != nil {
			return err
		}
		if err = binary.Write(buf, binary.LittleEndian, float32(val[2])); err != nil {
			return err
		}
	case *vector.Quaternion:
		if err = binary.Write(buf, binary.LittleEndian, float32(val.R)); err != nil {
			return err
		}
		if err = binary.Write(buf, binary.LittleEndian, float32(val.I)); err != nil {
			return err
		}
		if err = binary.Write(buf, binary.LittleEndian, float32(val.J)); err != nil {
			return err
		}
		if err = binary.Write(buf, binary.LittleEndian, float32(val.K)); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("Havent found out how to serialise literal %v with value of type '%T'", varType, val))
	}
	return err
}

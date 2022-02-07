package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/a-shine/butter"
	"github.com/a-shine/butter/node"
	"github.com/a-shine/butter/utils"
	uuid "github.com/nu7hatch/gouuid"
)

/* ---- GROUP ---- */

type Group struct {
	id      uuid.UUID
	members []utils.SocketAddr
	dataID  uuid.UUID
}

func (group *Group) ID() uuid.UUID {
	return group.id
}

func (group *Group) DataID() uuid.UUID {
	return group.dataID
}

func (group *Group) Members() *[]utils.SocketAddr {
	return &group.members
}

/* ---- PEER ---- */

type Peer struct {
	node   *node.Node
	groups []Group
}

func (peer *Peer) GetNode() *node.Node {
	return peer.node
}

func (peer *Peer) Groups() *[]Group {
	return &peer.groups
}

func (peer *Peer) Spawn() {

}

/* ---- NODE FUNCTIONALITY ---- */

func heartbeatEndpoint(_ *node.Node, payload []byte) []byte {
	message := string(payload)
	fmt.Println(message)
	return []byte("I'm alive too!")
}

func clientBehaviour(appInterface interface{}) {
	peer := appInterface.(*Peer)
	fmt.Println(len(peer.GetNode().KnownHosts()))
	for {
		Heartbeat(peer)
		time.Sleep(time.Second * 10)
	}
}

/* For all known hosts ping their heartbeat endpoint to inform them of our status.
 */
func Heartbeat(peer *Peer) {

	var recipients = make([]utils.SocketAddr, 0)

	recipients = peer.GetNode().KnownHosts()

	for _, group := range *peer.Groups() {
		for _, member := range *group.Members() {
			recipients = append(recipients, member)
		}
	}

	fmt.Println(len(recipients))
	//TODO only send to those we are in groups with.

	for i := 0; i < len(recipients); i++ { // For each known host
		response, err := utils.Request(recipients[i], []byte("heartbeat/"), []byte("I'm alive. Are you still alive?")) // Uses the utils package (recommended)
		if err != nil {
			// If there is an error, log the error BUT DO NOT FAIL - in decentralised application we avoid fatal
			// errors at all costs as we want to maximise node availability
			fmt.Println("Unable to send message to", recipients[i])
		}
		if bytes.Equal(response, []byte("I'm alive too!")) {
			fmt.Println(recipients[i], "is okay")
		}
		fmt.Println(recipients[i].ToString(), "responded with:", string(response))
	}
}

func CreatePeer() *Peer {

	peerNode, err := node.NewNode(0, 2048, clientBehaviour, false)
	if err != nil {
		fmt.Println("Fucked it")
		os.Exit(2)
	}

	return &Peer{&peerNode, make([]Group, 0)}
}

func main() {
	fmt.Println("hello world!")

	peer := CreatePeer()

	fmt.Println("Node is listening at", peer.GetNode().Address())

	// Specifying app level server behaviours - you can specify as many as you like as long as they are not reserved by
	// other butter packages
	peer.GetNode().RegisterRoute("heartbeat/", heartbeatEndpoint) // The client behaviour interacts with this route

	// Spawn your node into the butter network
	butter.Spawn(peer.GetNode(), false, peer) // Blocking

}

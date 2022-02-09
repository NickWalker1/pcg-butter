package main

import (
	"bufio"
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

// send a message to a specified host via the application specified reverse-message/ route
func send(remoteHost utils.SocketAddr, msg string) (string, error) {
	response, err := utils.Request(remoteHost, []byte("reverse-message/"), []byte(msg)) // Uses the utils package (recommended)
	if err != nil {
		return "", err
	}
	return string(response), nil
}


// Takes as input a string and returns the string in reverse.
func reverse(s string) string {
	rns := []rune(s) // Convert string to rune array
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		// Swap the letters of the string
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}

func heartbeatEndpoint(_ *node.Node, payload []byte) []byte {
	message := string(payload)
	fmt.Println(message)
	return []byte("I'm alive too!")
}

// The serverBehavior for this application is to reverse the packet it receives and return it back to the sender as a
// response
func revStrServ(_ *node.Node, payload []byte) []byte {
	message := string(payload)
	reversedMsg := reverse(message)
	return []byte(reversedMsg)
}


func clientBehaviour(appInterface interface{}) {
	peer := appInterface.(*Peer)
	fmt.Println(len(peer.GetNode().KnownHosts()))
	go func() {
		for {
			Heartbeat(peer)
			time.Sleep(time.Second * 5)
		}
	}()

	for {
		fmt.Print("Type message:")
		in := bufio.NewReader(os.Stdin)
		line, _ := in.ReadString('\n') // Read string up to newline

		knownHosts := peer.node.KnownHosts() // Get the node's known hosts

		for i := 0; i < len(knownHosts); i++ { // For each known host
			res, err := send(knownHosts[i], line) // Ask them to reverse the input message
			if err != nil {
				// If there is an error, log the error BUT DO NOT FAIL - in decentralised application we avoid fatal
				// errors at all costs as we want to maximise node availability
				fmt.Println("Unable to send message to", knownHosts[i])
			}
			fmt.Println(knownHosts[i].ToString(), "responded with:", res)
		}
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
			fmt.Println("Appears to have Died", recipients[i])
			peer.GetNode().KnownHosts()
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
	peer.GetNode().RegisterRoute("reverse-message/", revStrServ) // The client behaviour interacts with this route

	// Spawn your node into the butter network
	butter.Spawn(peer.GetNode(), false, peer) // Blocking

}

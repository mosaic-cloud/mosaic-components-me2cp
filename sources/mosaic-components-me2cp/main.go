

package main


import "fmt"
import "io"
import "net"
import "os"
import "sync"

import "vgl/transcript"


func main () () {
	
	var _controllerEndpointUrl = defaultControllerEndpointUrl
	var _channelEndpointIp string = defaultChannelEndpointIp
	var _channelEndpointPort uint16 = defaultChannelEndpointPort
	var _channelInboundStream *os.File = os.Stdin
	var _channelOutboundStream *os.File = os.Stdout
	
	_transcript := packageTranscript
	
	_transcript.TraceInformation ("initializing the ME2-based component proxy...")
	
	_transcript.TraceInformation ("preparing the channel endpoint accepter...")
	var _accepter net.Listener
	if _accepter_1, _error := net.Listen ("tcp", fmt.Sprintf ("%s:%d", _channelEndpointIp, _channelEndpointPort)); _error != nil {
		panic (_error)
	} else {
		_accepter = _accepter_1
	}
	if _accepterEndpoint, _ok := _accepter.Addr () .(*net.TCPAddr); !_ok {
		panic ("assertion")
	} else {
		_channelEndpointIp = _accepterEndpoint.IP.String ()
		_channelEndpointPort = uint16 (_accepterEndpoint.Port)
	}
	_transcript.TraceInformation ("  * using the endpoint `%s:%d`;", _channelEndpointIp, _channelEndpointPort)
	
	_transcript.TraceInformation ("preparing the the ME2 container...")
	_transcript.TraceInformation ("  * using the endpoint `%s`;", _controllerEndpointUrl)
	
	_transcript.TraceInformation ("waiting for a channel connection...")
	var _connection net.Conn
	if _connection_1, _error := _accepter.Accept (); _error != nil {
		panic (_error)
	} else {
		_connection = _connection_1
	}
	
	_transcript.TraceInformation ("conveying the channel messages...")
	var _waiting sync.WaitGroup
	_waiting.Add (1)
	go func () () {
		if _, _error := io.Copy (_connection, _channelInboundStream); _error != nil {
			panic (_error)
		}
		_transcript.TraceInformation ("closed inbound pipe...")
		_waiting.Done ()
	} ()
	go func () () {
		if _, _error := io.Copy (_channelOutboundStream, _connection); _error != nil {
			panic (_error)
		}
		_transcript.TraceInformation ("closed outbound pipe...")
		_waiting.Done ()
	} ()
	_waiting.Wait ()
	
	_transcript.TraceInformation ("terminated.")
}


var packageTranscript = transcript.NewPackageTranscript ()

var defaultControllerEndpointUrl = "http://127.0.0.1:xxx/"
var defaultChannelEndpointIp = "127.0.0.1"
var defaultChannelEndpointPort uint16 = 24704

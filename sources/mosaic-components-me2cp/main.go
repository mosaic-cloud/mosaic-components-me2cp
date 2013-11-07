

package main


import "fmt"
import "io"
import "net"
import "os"
import "strconv"
import "sync"


import "vgl/transcript"


func main () () {
	
	_transcript := packageTranscript
	_transcript.TraceInformation ("initializing the ME2-based component proxy...")
	
	var _componentIdentifier string
	var _channelEndpointIp string
	var _channelEndpointPort uint16
	var _channelInboundStream *os.File
	var _channelOutboundStream *os.File
	var _controllerUrl string
	var _bundle string
	
	_arguments := os.Args
	if len (_arguments) < 1 {
		_transcript.TraceError ("invalid arguments (expected at least one); aborting!")
		os.Exit (1)
	}
	switch _arguments[1] {
		
		case "component" :
			if ! (len (_arguments) == 4 || len (_arguments) == 5 || len (_arguments) == 7) {
				_transcript.TraceError ("invalid component arguments (expected 4, 5 or 7); aborting!")
				os.Exit (1)
			}
			_componentIdentifier = _arguments[2]
			_bundle = _arguments[3]
			if len (_arguments) >= 5 {
				_controllerUrl = _arguments[4]
			} else {
				_controllerUrl = defaultControllerUrl
			}
			if len (_arguments) >= 7 {
				_channelEndpointIp = _arguments[5]
				if _port, _error := strconv.ParseUint (_arguments[6], 10, 16); _error != nil {
					_transcript.TraceError ("invalid channel edpoint ip; aborting!")
					os.Exit (1)
				} else {
					_channelEndpointPort = uint16 (_port)
				}
			} else {
				_channelEndpointIp = os.Getenv ("mosaic_node_ip")
				if _channelEndpointIp == "" {
					_transcript.TraceError ("missing channel endpoint ip; aborting!")
					os.Exit (1)
				}
				_channelEndpointPort = 0
			}
			_channelInboundStream = os.Stdin
			_channelOutboundStream = os.Stdout
		
		case "standalone" :
			if len (_arguments) != 2 {
				_transcript.TraceError ("invalid standalone arguments (expected no others)")
				os.Exit (1)
			}
			_transcript.TraceError ("although standalone stdio is still piped...")
			_componentIdentifier = ""
			_channelEndpointIp = defaultChannelEndpointIp
			_channelEndpointPort = defaultChannelEndpointPort
			_controllerUrl = defaultControllerUrl
			_bundle = os.Getenv ("_me2cp_bundle")
			if _bundle == "" {
				_transcript.TraceError ("missing bundle; aborting!")
				os.Exit (1)
			}
			_channelInboundStream = os.Stdin
			_channelOutboundStream = os.Stdout
		
		default :
			_transcript.TraceError ("invalid component mode `%s`; aborting!", _arguments[1])
			os.Exit (1)
	}
	
	if _componentIdentifier != "" {
		_transcript.TraceInformation ("  * using the identifier `%s`;", _componentIdentifier)
	}
	
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
	
	_transcript.TraceInformation ("preparing the the ME2 controller...")
	_transcript.TraceInformation ("  * using the endpoint `%s`;", _controllerUrl)
	_transcript.TraceInformation ("  * using the bundle `%s`;", _bundle)
	
	_transcript.TraceInformation ("creating the the ME2 container...")
	_transcript.TraceInformation ("!!! not implemented yet !!!!")
	
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

var defaultControllerUrl = "http://127.0.0.1:8089/"
var defaultChannelEndpointIp = "127.0.0.1"
var defaultChannelEndpointPort uint16 = 24704

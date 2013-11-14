

package main


import "bytes"
import "encoding/json"
import "fmt"
import "io"
import "io/ioutil"
import "net"
import "net/http"
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
	var _configuration map[string]interface{}
	var _container string
	
	_arguments := os.Args
	if len (_arguments) < 1 {
		_transcript.TraceError ("invalid arguments (expected at least one); aborting!")
		os.Exit (1)
	}
	switch _arguments[1] {
		
		case "component" :
			if ! (len (_arguments) == 4 || len (_arguments) == 5 || len (_arguments) == 6 || len (_arguments) == 8) {
				_transcript.TraceError ("invalid component arguments (expected 4, 5, 6 or 8); aborting!")
				os.Exit (1)
			}
			_componentIdentifier = _arguments[2]
			_bundle = _arguments[3]
			if len (_arguments) >= 5 {
				if _error := json.Unmarshal ([]byte (_arguments[4]), &_configuration); _error != nil {
					_transcript.TraceError ("invalid configuration: `%s`; aborting!", _error.Error ())
					os.Exit (1)
				}
			}
			if len (_arguments) >= 6 {
				_controllerUrl = _arguments[5]
			} else {
				_controllerUrl = ""
			}
			if len (_arguments) >= 8 {
				_channelEndpointIp = _arguments[6]
				if _port, _error := strconv.ParseUint (_arguments[7], 10, 16); _error != nil {
					_transcript.TraceError ("invalid channel edpoint ip; aborting!")
					os.Exit (1)
				} else {
					_channelEndpointPort = uint16 (_port)
				}
			} else {
				_channelEndpointIp = os.Getenv (envkeyNodeIp)
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
				_transcript.TraceError ("invalid standalone arguments (expected no others); aborting!")
				os.Exit (1)
			}
			_transcript.TraceError ("standalone is not implemented; aborting!")
			os.Exit (1)
		
		default :
			_transcript.TraceError ("invalid component mode `%s`; aborting!", _arguments[1])
			os.Exit (1)
	}
	
	if _componentIdentifier != "" {
		_transcript.TraceInformation ("  * using the identifier `%s`;", _componentIdentifier)
	}
	
	if _controllerUrl == "" {
		_controllerUrl = os.Getenv (envkeyControllerUrl)
	}
	if _controllerUrl == "" {
		_controllerUrl = defaultControllerUrl
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
	
	_transcript.TraceInformation ("preparing the ME2 controller...")
	_transcript.TraceInformation ("  * using the endpoint `%s`;", _controllerUrl)
	_transcript.TraceInformation ("  * using the bundle `%s`;", _bundle)
	
	_transcript.TraceInformation ("creating the ME2 container...")
	var _startOutputs map[string]interface{}
	_startInputs := map[string]interface{} {
			"jsonrpc" : "2.0",
			"method" : "manager.start",
			"params" : map[string]interface{} {
				"bundle_id" : _bundle,
				"config" : map[string]interface{} {
						"component-identifier" : _componentIdentifier,
						"channel-endpoint" : fmt.Sprintf ("tcp:%s:%d", _channelEndpointIp, _channelEndpointPort),
						"configuration" : _configuration,
				},
			},
			"id" : 1,
	}
	if _startInputsData, _error := json.Marshal (_startInputs); _error != nil {
		panic (_error)
	} else if _response, _error := http.Post (_controllerUrl, "application/json", bytes.NewBuffer (_startInputsData)); _error != nil {
		panic (_error)
	} else if _startOutputsData, _error := ioutil.ReadAll (_response.Body); _error != nil {
		panic (_error)
	} else if _error := json.Unmarshal (_startOutputsData, &_startOutputs); _error != nil {
		_transcript.TraceError ("  * invalid response: `%s`; `%s;`", string (_startOutputsData), _error.Error ())
		panic (_error)
	} else {
		if _error, _exists := _startOutputs["error"]; _exists {
			_transcript.TraceError ("  * failed: `%#v`;", _error)
			panic ("failed")
		} else if _result_1, _exists := _startOutputs["result"]; !_exists {
			_transcript.TraceError ("  * invalid response: `%s`;", string (_startOutputsData))
			panic ("failed")
		} else if _result, _ok := _result_1.(string); !_ok {
			_transcript.TraceError ("  * invalid response: `%s`;", string (_startOutputsData))
			panic ("failed")
		} else {
			_container = _result
		}
		_transcript.TraceInformation ("  * started container: `%s`;", _container)
	}
	
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

var envkeyNodeIp = "mosaic_node_ip"
var envkeyControllerUrl = "mosaic_me2_controller_url"
var defaultControllerUrl = "http://127.0.0.1:8099/api"
var defaultStandaloneChannelEndpointIp = "127.0.0.1"
var defaultStandaloneChannelEndpointPort uint16 = 24704

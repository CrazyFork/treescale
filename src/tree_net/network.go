package tree_net
import (
	tree_path "tree_graph/path"
	"tree_event"
	"tree_node/node_info"
	"net"
	"bytes"
	"tree_log"
	"github.com/pquerna/ffjson/ffjson"
	"encoding/binary"
	"fmt"
	"tree_lib"
	"tree_graph"
	"tree_api"
)

var (
	api_connections		=	make(map[string]*net.TCPConn)
)

const (
	// This is just a random string, using to notify Parent or Child Node that one of them going to close connection
	CLOSE_CONNECTION_MARK = "***###***"
)


func init() {
	// Adding event emmit callback
	tree_event.NetworkEmitCB = NetworkEmmit
	tree_api.EmitApi	=	ApiEmit
	tree_api.EmitToApi	=	EmitToAPI
	// Child listener should be running without any condition
	go ChildListener(1000)
}

func Start() {
	ListenParent()
}

func Stop() {
	if parentConnection != nil {
		parentConnection.Close()
		parentConnection = nil
	}

	if listener != nil {
		listener.Close()
		listener = nil
	}

	for n, c :=range child_connections {
		if c != nil {
			c.Close()
		}
		delete(child_connections, n)
	}
}

func Restart() {
	Stop()
	Start()
}

func handle_message(is_api, from_parent bool, msg []byte) (err tree_lib.TreeError) {
	var (
		body_index	int
		path		tree_path.Path
		node_names	[]string
		handle_ev	bool
	)
	err.From = tree_lib.FROM_HANDLE_MESSAGE
	body_index, path, err = tree_path.PathFromMessage(msg)
	if !err.IsNull() {
		return
	}
	handle_ev = false


	if _, ok :=path.NodePaths[node_info.CurrentNodeInfo.Name]; !ok {
		handle_ev = true
	} else if _, _, ok :=tree_lib.ArrayMatchElement(path.Groups, node_info.CurrentNodeInfo.Groups); ok {
		handle_ev = true
	} else if _, _, ok :=tree_lib.ArrayMatchElement(path.Tags, node_info.CurrentNodeInfo.Tags); ok {
		handle_ev = true
	}

	if handle_ev {
		go tree_event.TriggerFromData(msg[body_index:])
	}

	if is_api {
		// If message came from API then it need's to be handled only on this node
		//	then if there would be path to send , handler will send it from event callback
		return
	}

	if from_parent {
		node_names = path.ExtractNames(node_info.CurrentNodeInfo, node_info.ChildsNodeInfo)
	} else {
		snf := node_info.ChildsNodeInfo
		snf[node_info.ParentNodeInfo.Name] = node_info.ParentNodeInfo
		node_names = path.ExtractNames(node_info.CurrentNodeInfo, snf)
	}

	err = SendToNames(msg[body_index:], &path, node_names...)

	return
}

func SendToNames(data []byte, path *tree_path.Path, names...string) (err tree_lib.TreeError) {
	err.From = tree_lib.FROM_SEND_TO_NAMES
	for _, n :=range names {
		var send_conn *net.TCPConn
		send_conn = nil

		if n_conn, ok := child_connections[n]; ok && n_conn != nil {
			send_conn = n_conn
		} else {
			if parent_name == n && parentConnection != nil {
				send_conn = parentConnection
			} else {
				if api_conn, ok := api_connections[n]; ok && api_conn != nil {
					send_conn = api_conn
				}
			}
		}

		if send_conn != nil {
			err = SendToConn(data, path, send_conn)
			if !err.IsNull() {
				tree_log.Error(err.From, err.Error())
			}
		}
	}
	return
}

func SendToConn(data []byte, path *tree_path.Path, conn *net.TCPConn) (err tree_lib.TreeError) {
	err.From = tree_lib.FROM_SEND_TO_CONN
	var (
		p_data	[]byte
		p_len	=	make([]byte, 4)
		msg_len	=	make([]byte, 4)
		buf		=	bytes.Buffer{}
	)
	p_data, err.Err = ffjson.Marshal(path)
	binary.LittleEndian.PutUint32(p_len, uint32(len(p_data)))
	binary.LittleEndian.PutUint32(msg_len, uint32(len(p_data)) + uint32(len(data)) + uint32(4))

	buf.Write(msg_len)
	buf.Write(p_len)
	buf.Write(p_data)
	buf.Write(data)

	_, err.Err = conn.Write(buf.Bytes())
	return
}

func TcpConnect(ip string, port int) (conn *net.TCPConn, err tree_lib.TreeError) {
	var (
		tcpAddr *net.TCPAddr
	)
	err.From = tree_lib.FROM_TCP_CONNECT
	tcpAddr, err.Err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port))
	if !err.IsNull() {
		return
	}

	conn, err.Err = net.DialTCP("tcp", nil, tcpAddr)
	return
}

func NetworkEmmit(em *tree_event.EventEmitter) (err tree_lib.TreeError) {
	var (
		path 		*tree_path.Path
		ev		=	em.Event
		sdata		[]byte
	)
	err.From = tree_lib.FROM_NETWORK_EMIT
	if len(em.Path.Groups) == 0 && len(em.Path.NodePaths) == 0 && len(em.Path.Tags) == 0 {
		path, err = tree_graph.GetPath(node_info.CurrentNodeInfo.Name, em.ToNodes, em.ToTags, em.ToGroups)
		if !err.IsNull() {
			return
		}

		if len(em.ToApi) > 0 && len(em.ToNodes) == 1 {
			ev.Path.NodePaths[em.ToNodes[0]] = em.ToApi
		}

		ev.Path = (*path)
	}

	// If from not set, setting it before network sending
	if len(em.From) == 0 {
		em.From = node_info.CurrentNodeInfo.Name
	}

	sdata, err.Err = ffjson.Marshal(ev)
	if !err.IsNull() {
		return
	}

	err = SendToNames(sdata, path, path.NodePaths[node_info.CurrentNodeInfo.Name]...)
	return
}

func ApiEmit(e *tree_event.Event, nodes...string) (err tree_lib.TreeError) {
	var (
		sdata	[]byte
	)
	err.From = tree_lib.FROM_API_EMIT
	// If from not set, setting it before network sending
	if len(e.From) == 0 {
		e.From = node_info.CurrentNodeInfo.Name
	}


	sdata, err.Err = ffjson.Marshal(e)
	if !err.IsNull() {
		return
	}

	for _, n :=range nodes {
		if c, ok :=child_connections[n]; ok && c != nil {
			err = SendToConn(sdata, &e.Path, c)
			if !err.IsNull() {
				tree_log.Error(err.From, "Unable to send data to node <", n, "> ", err.Error())
			}
		} else {
			tree_log.Error(err.From, "Please connect to Node <", n, "> before sending data")
		}
	}

	return
}

func EmitToAPI(e *tree_event.Event, apis...string) (err tree_lib.TreeError) {
	var (
		sdata	[]byte
	)
	err.From = tree_lib.FROM_EMIT_TO_API
	// If from not set, setting it before network sending
	if len(e.From) == 0 {
		e.From = node_info.CurrentNodeInfo.Name
	}


	sdata, err.Err = ffjson.Marshal(e)
	if !err.IsNull() {
		return
	}

	for _, n :=range apis {
		if c, ok :=api_connections[n]; ok && c != nil {
			err = SendToConn(sdata, &tree_path.Path{}, c)
			if !err.IsNull() {
				tree_log.Error(err.From, "Unable to send data to api <", n, "> ", err.Error())
			}
		} else {
			tree_log.Error(err.From, "Please connect to api <", n, "> before sending data")
		}
	}

	return
}
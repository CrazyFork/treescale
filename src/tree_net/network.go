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
)

var (
	api_connections		=	make(map[string]*net.TCPConn)
)

const (
	// This is just a random string, using to notify Parent or Child Node that one of them going to close connection
	CLOSE_CONNECTION_MARK = "***###***"
)


func init() {
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

func handle_message(is_api, from_parent bool, msg []byte) (err error) {
	var (
		body_index	int
		path		tree_path.Path
		node_names	[]string
		handle_ev	bool
	)
	body_index, path, err = tree_path.PathFromMessage(msg)
	if err != nil {
		return
	}
	handle_ev = false



	if p, ok :=path.NodePaths[node_info.CurrentNodeInfo.Name]; ok && len(p) == 0 {
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

func SendToNames(data []byte, path *tree_path.Path, names...string) (err error) {
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
			if err != nil {
				tree_log.Error("Send to Node Names", err.Error())
			}
		}
	}
	return
}

func SendToConn(data []byte, path *tree_path.Path, conn *net.TCPConn) (err error) {
	var (
		p_data	[]byte
		p_len	=	make([]byte, 4)
		msg_len	=	make([]byte, 4)
		buf		=	bytes.Buffer{}
	)
	p_data, err = ffjson.Marshal(path)
	binary.LittleEndian.PutUint32(p_len, uint32(len(p_data)))
	binary.LittleEndian.PutUint32(msg_len, uint32(len(p_data)) + uint32(len(data)))

	buf.Write(msg_len)
	buf.Write(p_len)
	buf.Write(p_data)
	buf.Write(data)

	_, err = conn.Write(buf.Bytes())
	return
}

func TcpConnect(ip string, port int) (conn *net.TCPConn, err error) {
	var (
		tcpAddr *net.TCPAddr
	)
	tcpAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return
	}

	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	return
}
package path

import (
	"encoding/binary"
	"github.com/pquerna/ffjson/ffjson"
	"tree_node/node_info"
	"tree_lib"
)

type Path struct {
	NodePaths		map[string][]string		`json:"node_paths" toml:"node_paths" yaml:"node_paths"`
	Groups			[]string				`json:"groups" toml:"groups" yaml:"groups"`
	Tags			[]string				`json:"tags" toml:"tags" yaml:"tags"`
}

func PathFromMessage(msg []byte) (body_index int, p Path, err tree_lib.TreeError) {
	// First 4 bytes in message is a length of json encoded Path
	err.From = tree_lib.FROM_PATH_FROM_MESSAGE
	path_len := int(binary.LittleEndian.Uint32(msg[:4]))
	err.Err = ffjson.Unmarshal(msg[4:path_len+4], &p)
	body_index = 4 + path_len
	return
}

func (path *Path) ExtractNames(current_node node_info.NodeInfo, nodes_info map[string]node_info.NodeInfo) (p_nodes []string) {
	var (
		ok			bool
	)

	if p_nodes, ok = path.NodePaths[current_node.Name]; ok {
		// deleting this path after getting it
		delete(path.NodePaths, current_node.Name)
	}

	for name, info :=range nodes_info {
		contains := false

		if _, _, ok :=tree_lib.ArrayMatchElement(path.Groups, info.Groups); ok {
			contains = true
		} else if _, _, ok := tree_lib.ArrayMatchElement(path.Tags, info.Tags); ok {
			contains = true
		}

		if contains {
			// If there is no duplicates
			if _, ok :=tree_lib.ArrayContains(p_nodes, name); !ok {
				p_nodes = append(p_nodes, name)
			}
		}
	}

	return
}
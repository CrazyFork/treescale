package tree_graph

import (
	"tree_db"
	"tree_node/node_info"
)

func GroupPath(group_name string) (map[string][]string, error){
	var (
		path = 			make(map[string][]string)
		err 			error
		nodes_in_group	[]string
	)
	nodes_in_group, err = tree_db.GetGroupNodes(group_name)
	if err != nil {
		return nil, err
	}
	path, err = NodePath(nodes_in_group[0])
	if err != nil {
		return nil, err
	}
	return path, nil
}

func NodePath(node_name string) (map[string][]string, error){
	var (
		node					string
		err						error
		path =					make(map[string][]string)
		path1 					[]string
		relations = 			make(map[string][]string)
		nodes  					[]string
		from = 					make(map[string]string)
	)
	nodes, err = tree_db.ListNodeNames()
	if err != nil {
		return nil, err
	}
	for _, a := range nodes {
		relations[a], err = tree_db.GetRelations(a)
		if err != nil {
			return nil, err
		}
	}

	from = bfs(node_name, relations)
	node = node_info.CurrentNodeInfo.Name
	for node != node_name {
		path1 = append(path1, node)
		node = from[node]
	}
	path1 = append(path1, node_info.CurrentNodeInfo.Name)

	for i := len(path1)-1; i>0; i-- {
		path[path1[i]] = append(path[path1[i]], path1[i-1])
	}

	return path, nil
}

func TagPath(tag_name string) (map[string][]string, error){
	var (
		err						error
		path =					make(map[string][]string)
		nodes_by_tagname 		[]string
		paths =					make(map[string]map[string][]string)
	)
	nodes_by_tagname, err = tree_db.GetNodesByTagName(tag_name)
	if err != nil {
		return nil, err
	}
	for _, a := range nodes_by_tagname {
		paths[a], err = NodePath(a)
		if err != nil {
			return nil, err
		}
	}
	path = merge(paths, nil, nil)
	return path, nil
}

func merge(nodes_path map[string]map[string][]string, groups_path map[string]map[string][]string, tags_path map[string]map[string][]string) (path map[string][]string){
	for _, a := range nodes_path {
		for i, b := range a{
			for _, c := range b {
				path[i] = append(path[i], c)
			}
		}
	}
	if groups_path != nil {
		for _, a := range groups_path {
			for i, b := range a{
				for _, c := range b {
					path[i] = append(path[i], c)
				}
			}
		}
	}
	if tags_path != nil {
		for _, a := range tags_path {
			for i, b := range a{
				for _, c := range b {
					path[i] = append(path[i], c)
				}
			}
		}
	}
	return
}

func bfs(end string, nodes map[string][]string) map[string]string{
	frontier := []string{node_info.CurrentNodeInfo.Name}
	visited := map[string]bool{}
	next := []string{}
	from := map[string]string{}

	for 0 < len(frontier) {
		next = []string{}
		for _, node := range frontier {
			visited[node] = true
			if node == end {
				return from
			}
			for _, n := range bfs_frontier(node, nodes, visited) {
				next = append(next, n)
				from[n] = node
			}
		}
		frontier = next
	}
	return nil
}

func bfs_frontier(node string, nodes map[string][]string, visited map[string]bool) []string {
	next := []string{}
	iter := func (n string) bool { _, ok := visited[n]; return !ok }
	for _, n := range nodes[node] {
		if iter(n) {
			next = append(next, n)
		}
	}
	return next
}

func GetPath(nodes []string, tags []string, groups []string) (*Path, error){
	var (
		err					error
		path 				*Path
		nodes_path =		make(map[string]map[string][]string)
		tags_path =			make(map[string]map[string][]string)
		groups_path = 		make(map[string]map[string][]string)
		final_path =		make(map[string][]string)
	)
	for _, a := range nodes {
		nodes_path[a], err = NodePath(a)
		if err != nil {
			return nil, err
		}
	}
	for _, a := range groups {
		groups_path[a], err = GroupPath(a)
		if err != nil {
			return nil, err
		}
	}
	for _, a := range tags {
		tags_path[a], err = TagPath(a)
		if err != nil {
			return nil, err
		}
	}
	final_path = merge(nodes_path, groups_path, tags_path)
	path.NodePaths = final_path
	path.Tags = tags
	path.Groups = groups
	return path, nil
}
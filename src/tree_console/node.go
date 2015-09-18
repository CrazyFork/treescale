package tree_console
import (
	"github.com/spf13/cobra"
	"tree_node"
	"tree_log"
	"fmt"
	"tree_db"
	"tree_event"
	"time"
)

const (
	log_from_node_console = "Console functionality for Node"
)

func HandleNodeCommand(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		tree_log.Error(log_from_node_console, err.Error())
		return
	}

	if len(name) == 0 {
		current_node_byte, err := tree_db.Get(tree_db.DB_RANDOM, []byte("current_node"))
		if err != nil {
			tree_log.Error(log_from_node_console, "Getting current node name from Random database, ", err.Error())
			return
		}
		if len(current_node_byte) == 0 {
			fmt.Println("Name is important for the first time run")
			return
		}
	} else {
		err = tree_node.SetCurrentNode(name)
		if err != nil {
			tree_log.Error(log_from_node_console, err.Error())
			return
		}
	}

	tree_event.ON("test", func(e *tree_event.Event) bool {
		fmt.Println(e.Data)
		return true
	})

	go func() {
		time.Sleep(time.Second * 2)
		if name == "tree1" {
			em := &tree_event.EventEmitter{}
			em.Name = "test"
			em.Data = []byte("aaaaaaaaaaaaaaaa")
			em.ToNodes = []string{"tree2"}
			err := tree_event.Emit(em)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	tree_node.Start()
}
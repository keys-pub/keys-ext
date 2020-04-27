package cmds

import (
	"encoding/json"
	"fmt"
)

func printResponse(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

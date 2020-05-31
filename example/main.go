package main

import (
	"fmt"

	"github.com/liadmire/sys"
)

func main() {
	fmt.Println(sys.SelfDir(), sys.SelfPath(), sys.SelfName(), sys.SelfExt())
}

package main

import (
	"oss.terrastruct.com/d2/d2plugin"
	"oss.terrastruct.com/util-go/xmain"

	"github.com/valllabh/octopus-layout-engine/internal/plugin"
)

func main() {
	xmain.Main(d2plugin.Serve(plugin.New()))
}

package main

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"

	"github.com/mach-composer/mach-composer-plugin-vercel/internal"
)

func main() {
	p := internal.NewVercelPlugin()
	plugin.ServePlugin(p)
}

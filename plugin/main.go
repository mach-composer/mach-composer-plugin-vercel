package plugin

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"

	"github.com/mach-composer/mach-composer-plugin-vercel/internal"
)

func Serve() {
	p := internal.NewVercelPlugin()
	plugin.ServePlugin(p)
}

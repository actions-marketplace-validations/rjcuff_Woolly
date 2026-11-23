package main

import (
	"fmt"
	"net/http"

	"github.com/rjcuff/woolly/internal/mcp"
	"github.com/rjcuff/woolly/internal/memory"
)

// mcpHTTP serves MCP over HTTP at /mcp on `woolly serve`. Each session can
// be scoped like `woolly mcp`, with ?user=, ?agent= and ?project= in the URL.
// There's no working folder over HTTP, so sources are dropped rather than
// saved unchecked, and the why and check_memory_files tools are left out.
func (a *app) mcpHTTP() *mcp.HTTPHandler {
	return &mcp.HTTPHandler{New: func(r *http.Request) (*mcp.Server, error) {
		q := r.URL.Query()
		for _, name := range []string{"user", "agent"} {
			if q.Has(name) && memory.CleanScopeID(q.Get(name)) == "" {
				return nil, fmt.Errorf("%w: %s is blank", mcp.ErrBadRequest, name)
			}
		}
		return &mcp.Server{
			Store:     a.store,
			Engine:    a.engine,
			Version:   version,
			Activity:  a.log,
			Project:   memory.NormalizeProject(q.Get("project")),
			User:      memory.CleanScopeID(q.Get("user")),
			Agent:     memory.CleanScopeID(q.Get("agent")),
			NoSources: true,
		}, nil
	}}
}

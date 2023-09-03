package qhttp

import (
	"net/http"
)

type Router struct {
	tree *routeTree
}

// Constructor for router
func HttpRouter() *Router {
	router_tree := createRouteTree()
	http_router := new(Router)
	http_router.tree = router_tree
	return http_router
}

// Create API Route
func (r *Router) Register(req_type string, path string, new_func func(res http.ResponseWriter, req *http.Request)) {
	r.tree.register(req_type, path, new_func)
}

func (r *Router) OpenPort(port string) {
	http.ListenAndServe(port, r.tree)
}

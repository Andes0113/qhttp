package qhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Node struct {
	path_name string
	// Functions corresponding to this path that
	registered_functions map[string]func(res http.ResponseWriter, req *http.Request)
	// Routes with paths that follow after node (e.g. /users => /users/settings)
	children map[string]*Node
}

func (n *Node) registerFunction(req_type string, path []string, new_func func(res http.ResponseWriter, req *http.Request)) error {
	if len(path) == 0 {
		// Register function
		(*n).registered_functions[req_type] = new_func
		return nil
	}
	// Find child whose path name matches next string in url path
	child, ok := (*n).children[path[0]]
	if !ok {
		// If path not created, create path
		new_node := new(Node)
		(*new_node).path_name = path[0]
		(*new_node).children = make(map[string]*Node)
		(*new_node).registered_functions = make(map[string]func(res http.ResponseWriter, req *http.Request))
		(*n).children[path[0]] = new_node
		child = new_node
	}
	// Search child node that follows in path
	return (*child).registerFunction(req_type, path[1:], new_func)
}

func (n *Node) findFunction(req_type string, path []string) (func(res http.ResponseWriter, req *http.Request), error) {
	if n == nil {
		return nil, errors.New("unregistered function")
	}
	// Empty path array means we've arrived to target route node
	if len(path) == 0 {
		// Get function from registered_functions
		matched_function, ok := (*n).registered_functions[req_type]
		if !ok {
			// If no matching function, throw error
			return nil, errors.New("unregistered function")
		}
		return matched_function, nil
	}
	child, ok := (*n).children[path[0]]
	if !ok {
		return nil, errors.New("unregistered function")
	}

	// Traverse child that follows in url path
	return (*child).findFunction(req_type, path[1:])
}

// Interface for easier tree traversal
type routeTree struct {
	Root *Node
}

// Convert request url path to format usable by route tree
func pathToArr(path string) []string {
	var path_arr []string
	if path[1:] != "" {
		path_arr = strings.Split(path[1:], "/")
	}
	return path_arr
}

// Add api route and function to routeTree
func (rt *routeTree) register(req_type string, path string, new_func func(res http.ResponseWriter, req *http.Request)) {
	rt.Root.registerFunction(req_type, pathToArr(path), new_func)
}

func (rt *routeTree) getRouteFunction(req_type string, path string) (func(res http.ResponseWriter, req *http.Request), error) {
	return rt.Root.findFunction(req_type, pathToArr(path))
}

func (r *routeTree) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Get function corresponding to url path
	route_function, err := r.getRouteFunction(req.Method, req.URL.Path)

	// If no route found, return http error
	if err != nil {
		http.Error(res, "No matching route found", http.StatusNotFound)
		return
	}

	// Call function for route if found
	route_function(res, req)
}

// Constructor function for route tree
func createRouteTree() *routeTree {
	// Create root node for RouteTree
	root := new(Node)
	// Recognizable path name for base node. Just for human readability
	(*root).path_name = "/"

	// Initialize rest of root node
	(*root).children = make(map[string]*Node)
	(*root).registered_functions = make(map[string]func(res http.ResponseWriter, req *http.Request))

	// Create RouteTree
	rt := new(routeTree)
	rt.Root = root

	return rt
}

// Debugging methods to view tree structure
func preorderTraverse(curr *Node) {
	fmt.Print((*curr).path_name)
	fmt.Println("\t", (*curr).registered_functions)
	for _, n := range (*curr).children {
		preorderTraverse(n)
	}
}
func (rt *routeTree) preorder() {
	preorderTraverse(rt.Root)
}

package qhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Node struct {
	path_name            string
	registered_functions map[string]func(res http.ResponseWriter, req *http.Request)
	children             map[string]*Node
}

func (n *Node) registerFunction(req_type string, path []string, new_func func(res http.ResponseWriter, req *http.Request)) error {
	if len(path) == 0 {
		// Register function
		(*n).registered_functions[req_type] = new_func
		return nil
	}
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
	return (*child).registerFunction(req_type, path[1:], new_func)
}

func (n *Node) findFunction(req_type string, path []string) (func(res http.ResponseWriter, req *http.Request), error) {
	if n == nil {
		return nil, errors.New("unregistered function")
	}
	if len(path) == 0 {
		// Get function from registered_functions
		matched_function, ok := (*n).registered_functions[req_type]
		if !ok {
			return nil, errors.New("unregistered function")
		}
		return matched_function, nil
	}
	child, ok := (*n).children[path[0]]
	if !ok {
		return nil, errors.New("unregistered function")
	}
	return (*child).findFunction(req_type, path[1:])
}

func PreorderTraverse(curr *Node) {
	fmt.Print((*curr).path_name)
	fmt.Println("\t", (*curr).registered_functions)
	for _, n := range (*curr).children {
		PreorderTraverse(n)
	}
}

type RouteTree struct {
	Root *Node
}

func pathToStack(path string) []string {
	var path_arr []string
	if path[1:] != "" {
		path_arr = strings.Split(path[1:], "/")
	}
	return path_arr
}

func (rt *RouteTree) register(req_type string, path string, new_func func(res http.ResponseWriter, req *http.Request)) {
	rt.Root.registerFunction(req_type, pathToStack(path), new_func)
}

func (rt *RouteTree) getRouteFunction(req_type string, path string) (func(res http.ResponseWriter, req *http.Request), error) {
	return rt.Root.findFunction(req_type, pathToStack(path))
}

func (r *RouteTree) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ret_func, err := r.getRouteFunction(req.Method, req.URL.Path)
	if err != nil {
		http.Error(res, "No matching route found", http.StatusNotFound)
		return
	}
	ret_func(res, req)
}

func createRouteTree() *RouteTree {
	// Create root node for RouteTree
	root := new(Node)
	(*root).path_name = "/"
	(*root).children = make(map[string]*Node)
	(*root).registered_functions = make(map[string]func(res http.ResponseWriter, req *http.Request))

	// Create RouteTree
	rt := new(RouteTree)
	rt.Root = root

	return rt
}

func (rt *RouteTree) preorder() {
	PreorderTraverse(rt.Root)
}

package gee

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

//roots 例子:roots['GET'] roots['POST']
//handlers 例子:handlers['GET-/p/:lang/doc'] , handlers['POST-/p/book']

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

//只允许一个*
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	fmt.Println(parts)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	fmt.Println(ok)
	if !ok {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(pattern, parts, 0)
	fmt.Println(r.roots[method].pattern,r.roots[method].part)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)//分割相对路径
	fmt.Println("search",searchParts)
	params := make(map[string]string)
	root, ok := r.roots[method]//root属于树结构
						//           /
						//        ↙    ↘
						//   /hello    assets
						//      ↙         ↘
						//    :name     *filepath
	fmt.Println("root",root.pattern,root.children[1].children[0])
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	fmt.Println("node",n)
	if n != nil {
		parts := parsePattern(n.pattern)
		fmt.Println("parts",parts)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
			fmt.Println("params",params)
		}
		return n, params
	}
	return nil, nil
}

func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	fmt.Println("yb test2",params,n)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

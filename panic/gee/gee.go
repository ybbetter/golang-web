package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

//处理请求函数
type HandlerFunc func(c *Context)

//一个用来处理分组的结构体
//一个最顶层的结构体
type (
	RouterGroup struct {
		prefix      string        //存放对应分组的前缀，用来查找此前缀是否有所支持的中间件
		middlewares []HandlerFunc //此分组支持的中间件
		parent      *RouterGroup
		engine      *Engine
	}
	Engine struct {
		*RouterGroup                     //继承RouterGroup的所有属性
		router        *router            //动态路由
		groups        []*RouterGroup     //存储所有的RouterGroup
		htmlTemplates *template.Template // 便于html文件渲染
		funcMap       template.FuncMap   //便于html文件渲染
	}
)

//构造函数
func New() *Engine {
	engine:=&Engine{router: newRouter()}//初始化engine
	engine.RouterGroup=&RouterGroup{engine: engine}//初始化
	engine.groups=[]*RouterGroup{engine.RouterGroup}
	return engine
}

//默认使用Logger()和异常恢复中间件
func Default() *Engine {
	engine:=New()
	engine.Use(Logger(),Recovery())
	return engine
}
//RouterGroup构造函数
func (group *RouterGroup)Group(prefix string) *RouterGroup {
	engine:=group.engine
	newGroup:=&RouterGroup{
		prefix:      group.prefix+prefix,
		parent:      group,
		engine:      engine,
	}
	engine.groups=append(engine.groups,newGroup)
	return newGroup
}
//向分组中添加对应中间件
func (group *RouterGroup)Use(middlewares ...HandlerFunc)  {
	group.middlewares=append(group.middlewares,middlewares...)
}
//添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}
// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}

// for custom render function
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

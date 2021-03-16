package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//定义一个可以映射任意类型的map
type H map[string]interface{}

type Context struct {
	//ServerHTTP所需参数
	Writer http.ResponseWriter//回复
	Req *http.Request//请求
	//请求的信息
	Path string//所访问的路由前缀如/hello、/admin
	Method string//请求使用的方法如GET、POST
	Params map[string]string//用户请求所携带的参数,eg:/hello?username=""&passwd=""
	//回复信息
	StatusCode int//状态码
	//中间件
	handlers []HandlerFunc//维护一个处理函数的数组，方便处理函数与对应中间件的映射
	index int//数组下标表示函数所在位置
	//engine指针
	engine *Engine//继承Engine的属性
}

//构造函数
func newContext(w http.ResponseWriter,req *http.Request) *Context {
	return &Context{
		Writer:     w,
		Req:        req,
		Path:       req.URL.Path,
		Method:     req.Method,
		index:      -1,
	}
}

//用来执行中间件的函数
func (c *Context)Next()  {
	c.index++
	s:=len(c.handlers)
	for ;c.index<s;c.index++{
		c.handlers[c.index](c)
	}
}

//处理失败逻辑
func (c *Context)Fail(code int,err string)  {
	c.index=len(c.handlers)
	c.JSON(code,H{"message":err})
}

//获取参数
func (c *Context)Param(key string) string {
	value,_:=c.Params[key]
	return value
}
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}
//设定写入的状态码和头部
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}
//封装回复的格式
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML template render
// refer https://golang.org/pkg/html/template/
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
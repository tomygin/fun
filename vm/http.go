package vm

import (
	"fmt"
	"fun/number"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// HTTP：客户端 + 服务端
//
// 设计哲学：一切都是表。
//   - 响应是表：{ ok, status, body, headers }
//   - 请求是表：处理函数收到 { method, path, query, headers, body }
//   - 路由是表：{ "/hello": fun(req){...}, "*": fun(req){...} }
//   - 处理函数就是普通的 Fun 函数，返回字符串或 { status, body, headers }
//
// 并发说明：Go 的 http 服务器天然并发，而 Fun 解释器是单线程协作式的，
// 所以所有处理函数由互斥锁串行执行（同一时刻只有一个 handler 在跑）。
// ----------------------------------------------------------------------------

// httpTable 构造暴露给脚本的 http 模块表（方法闭包持有 vm 以便回调解释器）
func (vm *VM) httpTable() map[string]any {
	return map[string]any{
		// http.get(url [, headers]) -> { ok, status, body, headers }
		"get": func(args ...any) any {
			if len(args) < 1 {
				return httpError("get requires a url")
			}
			headers, _ := argTable(args, 1)
			return httpDo("GET", formatValue(args[0]), "", headers, 30)
		},
		// http.post(url, body [, headers]) -> { ok, status, body, headers }
		"post": func(args ...any) any {
			if len(args) < 2 {
				return httpError("post requires url and body")
			}
			headers, _ := argTable(args, 2)
			return httpDo("POST", formatValue(args[0]), formatValue(args[1]), headers, 30)
		},
		// http.request({ method, url, body, headers, timeout }) 完全控制
		"request": func(args ...any) any {
			opts, ok := argTable(args, 0)
			if !ok {
				return httpError("request requires an options table")
			}
			method := "GET"
			if m, ok := opts["method"].(string); ok {
				method = strings.ToUpper(m)
			}
			urlStr, _ := opts["url"].(string)
			if urlStr == "" {
				return httpError("request requires url")
			}
			body := ""
			if b, ok := opts["body"]; ok && b != nil {
				body = formatValue(b)
			}
			headers, _ := opts["headers"].(map[string]any)
			timeout := 30.0
			if t, ok := opts["timeout"]; ok {
				if f, ok := number.ToFloat(t); ok && f > 0 {
					timeout = f
				}
			}
			return httpDo(method, urlStr, body, headers, timeout)
		},
		// http.serve(addr, routes) 前台启动服务器（阻塞，不返回）
		"serve": func(args ...any) any {
			if len(args) < 2 {
				return httpError("serve requires addr and routes")
			}
			addr := formatValue(args[0])
			err := http.ListenAndServe(addr, vm.buildHandler(args[1]))
			return httpError(err.Error())
		},
		// http.listen(addr, routes) 后台启动服务器（端口就绪后立即返回）
		"listen": func(args ...any) any {
			if len(args) < 2 {
				return httpError("listen requires addr and routes")
			}
			addr := formatValue(args[0])
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return httpError(err.Error())
			}
			go func() { _ = http.Serve(ln, vm.buildHandler(args[1])) }()
			return true
		},
	}
}

// argTable 取第 i 个参数并断言为表
func argTable(args []any, i int) (map[string]any, bool) {
	if i < len(args) {
		if m, ok := args[i].(map[string]any); ok {
			return m, true
		}
	}
	return nil, false
}

// httpError 构造错误响应表
func httpError(msg string) map[string]any {
	return map[string]any{"ok": false, "status": 0, "body": "", "error": msg}
}

// httpDo 执行一次 HTTP 请求，返回响应表
func httpDo(method, urlStr, body string, headers map[string]any, timeoutSec float64) map[string]any {
	client := &http.Client{Timeout: time.Duration(timeoutSec * float64(time.Second))}

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return httpError(err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, formatValue(v))
	}

	resp, err := client.Do(req)
	if err != nil {
		return httpError(err.Error())
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpError(err.Error())
	}

	hdrs := make(map[string]any, len(resp.Header))
	for k := range resp.Header {
		hdrs[k] = resp.Header.Get(k)
	}
	return map[string]any{
		"ok":      resp.StatusCode < 400,
		"status":  resp.StatusCode,
		"body":    string(data),
		"headers": hdrs,
	}
}

// buildHandler 把"路由表 或 单个处理函数"变成 Go 的 http.Handler。
// 路由表按 path 精确匹配，"*" 是兜底；单个函数处理所有路径。
func (vm *VM) buildHandler(routes any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := resolveRoute(routes, r.URL.Path)
		if handler == nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, "not found")
			return
		}

		// 组装请求表
		bodyBytes, _ := io.ReadAll(r.Body)
		query := make(map[string]any)
		for k, vs := range r.URL.Query() {
			if len(vs) > 0 {
				query[k] = vs[0]
			}
		}
		hdrs := make(map[string]any, len(r.Header))
		for k := range r.Header {
			hdrs[k] = r.Header.Get(k)
		}
		reqTable := map[string]any{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   query,
			"headers": hdrs,
			"body":    string(bodyBytes),
		}

		// Fun 解释器单线程：处理函数串行进入
		vm.httpMu.Lock()
		result, err := vm.applyFunction(handler, []any{reqTable})
		vm.httpMu.Unlock()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, err.Error())
			return
		}
		writeFunResponse(w, result)
	}
}

// resolveRoute 从路由里找处理函数
func resolveRoute(routes any, path string) any {
	rt, ok := routes.(map[string]any)
	if !ok {
		return nil
	}
	// 单个函数：处理全部路径
	if isUserFunc(rt) {
		return rt
	}
	if h, ok := rt[path]; ok {
		return h
	}
	if h, ok := rt["*"]; ok {
		return h
	}
	return nil
}

// writeFunResponse 把处理函数的返回值写成 HTTP 响应：
//
//	字符串        -> 200 + 正文
//	表           -> { status: 200, body: "", headers: {} } 三个字段都可选
//	nil          -> 200 空响应
//	其它（数字等） -> 200 + 格式化后的文本
func writeFunResponse(w http.ResponseWriter, result any) {
	switch res := result.(type) {
	case nil:
		w.WriteHeader(http.StatusOK)
	case string:
		_, _ = io.WriteString(w, res)
	case map[string]any:
		if hs, ok := res["headers"].(map[string]any); ok {
			for k, v := range hs {
				w.Header().Set(k, formatValue(v))
			}
		}
		status := http.StatusOK
		if s, ok := res["status"]; ok {
			if f, ok := number.ToFloat(s); ok && f > 0 {
				status = int(f)
			}
		}
		w.WriteHeader(status)
		if b, ok := res["body"]; ok && b != nil {
			_, _ = io.WriteString(w, formatValue(b))
		}
	default:
		_, _ = fmt.Fprint(w, formatValue(res))
	}
}

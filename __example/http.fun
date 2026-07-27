// ============================================================
// HTTP：客户端 + 服务端 + JSON —— 一切都是表
//
//   http.get(url [, headers])          -> { ok, status, body, headers }
//   http.post(url, body [, headers])   -> 同上
//   http.request({ method, url, body, headers, timeout })
//   http.serve(addr, routes)           前台启动（阻塞）
//   http.listen(addr, routes)          后台启动（立即返回）
//   json.encode(v) / json.decode(s)    表 <-> JSON
//
// 路由是表：path -> 处理函数，"*" 兜底；
// 处理函数收到请求表 { method, path, query, headers, body }，
// 返回字符串或 { status, body, headers }。
// ============================================================

print("========== 1. JSON：表与文本互转 ==========")

user := { name: "Gin", age: 18, tags: { "go", "lua" }, score: 0.1 }
s := json.encode(user)
print("encode:", s)

back := json.decode(s)
print("decode: name =", back.name, ", tags[0] =", back.tags[0])
// JSON 里的数字走数字塔,小数依然精确：
print("小数往返精确:", back.score + 0.2 == 0.3)

print("========== 2. 起一个服务器 ==========")

// 路由就是一张表；处理函数就是普通 Fun 函数（还是闭包！）
visits := 0
routes := {
    "/hello": fun(req) {
        return "hello, " + (req.query["name"] || "世界")
    },
    "/visit": fun(req) {
        visits = visits + 1              // 处理函数能读写外面的变量
        return `第 ${visits} 位访客`
    },
    "/user": fun(req) {
        return {
            status: 200,
            headers: { "Content-Type": "application/json" },
            body: json.encode({ name: "Gin", vip: true })
        }
    },
    "/echo": fun(req) {
        return { status: 201, body: req.method + ": " + req.body }
    },
    "*": fun(req) {
        return { status: 404, body: "没有这个路由: " + req.path }
    }
}

// listen 后台启动（本例要继续跑客户端）；真正的服务用 http.serve 阻塞前台
http.listen("127.0.0.1:18080", routes)
print("服务器已启动在 127.0.0.1:18080")

print("========== 3. 客户端把自己请求一遍 ==========")

r := http.get("http://127.0.0.1:18080/hello?name=Tom")
print("GET /hello  ->", r.status, "|", r.body)

http.get("http://127.0.0.1:18080/visit")
r = http.get("http://127.0.0.1:18080/visit")
print("GET /visit  ->", r.body, "（脚本变量 visits =", visits, "）")

r = http.get("http://127.0.0.1:18080/user")
u := json.decode(r.body)
print("GET /user   ->", u.name, ", vip =", u.vip,
      ", Content-Type =", r.headers["Content-Type"])

r = http.post("http://127.0.0.1:18080/echo", "ping!")
print("POST /echo  ->", r.status, "|", r.body)

r = http.get("http://127.0.0.1:18080/nothing")
print("GET 未知路由 ->", r.status, "| ok =", r.ok, "|", r.body)

// request 完全体：方法/头/超时自己定
r = http.request({
    method: "GET",
    url: "http://127.0.0.1:18080/hello?name=Request",
    headers: { "X-Client": "fun" },
    timeout: 5
})
print("http.request ->", r.body)

print("HTTP 演示结束 🌐")

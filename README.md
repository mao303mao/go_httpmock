# go_httpmock
  - 这个是一个简单的http(s) proxy方式的mock工具,基于包"github.com/elazarl/goproxy"
  - 支持url转发
  - 支持更新响应的头及内容
  - 支持直接返回新的响应
  - 规则文件可以随时更改，每10s更新一次
  - 支持设置上行http代理，启动时生效

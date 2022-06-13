# go_http_proxy_mock
  - 这个是一个简单的http(s) proxy方式的mock工具,基于包"github.com/elazarl/goproxy"
  - 支持url转发
  - 支持更新响应的头及内容
  - 支持直接返回新的响应
  - 规则文件可以随时更改，每10s更新一次
  - 支持设置上行http代理，启动时生效


# 【一】添加证书为根路径信任证书
windows安装z.x509.cer证书（不小心删除也没关系会重新生成），选择第三方根证书颁发机构，这样重启浏览器，使用go_httpmock的代理时的https请求就可信任了。
如果是其他系统平台，可以修改cert.go中的代码+百度/Google

# 【二】配置上行（upstream）代理
在启动之前配置upstreamProxyConfig.json即可设置上行代理，结构如下：
```
{
  "proxyActive": false,
  "proxyUrl":"http://192.168.16.67:8080",
  "proxyUser": "",
  "proxyPassword": ""
}
```
- proxyActive： false表示不使用上行代理，true表示启用
- proxyUrl：如上格式，表示使用67代理，如想转发到类似fiddler上，可以配置"http://127.0.0.1:8888"
- proxyUser，proxyPassword：则表示上行代理需要验证的情况，输入对应用户名、密码

# 【三】方法一（推荐用法）：访问127.0.0.1:8088修改规则
  原理同“【四】方法二：代理规则（rules.json）说明”
  页面上提供了jsoneditor加上json-schema方式来修改rules.json
 ![image](https://user-images.githubusercontent.com/37785668/173273896-2639999c-b2dd-45fc-af62-7079da7e72af.png)

# 【四】方法二(不建议直接使用)：代理规则（rules.json）说明
规则文件rules.json可随意更改，每隔10s自动更新一次（控制台中有提示）
有2种响应mock，1种是新建http响应，另1种是修改http响应。

## 新建http响应
这种不要求服务端可用，使用构造响应或转发url来返回相应，对应字段"newRespRules"，它是个规则的列表。

### 【1】构造响应的规则的结构
```
{
      "active": true,
      "urlMatchRegexp": "/api/v1/account/getUserInfo",
      "respAction": {
        "setHeaders": [
          {"header": "Access-Control-Allow-Origin","value": "https://www.xxx.com"},
          {"header": "Set-Cookie","value": "lui=VjZnM1N0eGlYQnNZVlNjeTJHWjI0UT09;path=/;domain=.xxx.com;HttpOnly"}
        ],

         "bodyFile": "./respFiles/getUserInfo1.json"
        
      }
    }
```
- active：false表示规则禁用，true表示规则启用
- urlMatchRegexp：表示url匹配的正则表示式(注意json中"\"要改成"\\")
- respAction：包含setHeaders和setBody
- setBody：因为是是构造响应，bodyFile。填写对应响应文件的路径（比如正常情况都放在respFiles文件夹下）
- setHeaders：是设置相应头的规则列表，比如这里设置了2个响应头Access-Control-Allow-Origin和设置cookie。如不想设置，"setHeaders":[] 或"setHeaders":null

### 【2】转发url的规则的结构
转发url 折叠源码
```
{
      "active": false,
      "urlMatchRegexp": "(channelhub/api/v1/shoppingCart/list)",
      "reWriteUrl": "http://www.baidu.com/${1}",
      "respAction":null
 }
 ```
- active：false表示规则禁用，true表示规则启用
- urlMatchRegexp：表示url匹配的正则表示式(注意json中"\"要改成"\\")，如用(...)表示这里可以获取子匹配，可以用于reWriteUrl中
- reWriteUrl：转发的url，可以使用urlMatchRegexp的子匹配(${1}、${2}...)，比如这里表示转发到"http://www.baidu.com/channelhub/api/v1/shoppingCart/list"
- respAction：当reWriteUrl有内容时，respAction就没有用了，这里设置null即可

## 修改http响应
这种要求服务端可用，使用规则来更新服务端返回的响应的头及内容，对应字段"updateRespRules"，它是个规则的列表。
更新响应的规则的结构
```
{
      "active": true,
      "urlMatchRegexp": "/api/v1/account/getUserInfo",
      "respAction": {
        "setHeaders": [
          {"header": "Access-Control-Allow-Origin","value": "https://www.xxx.com"},
          {"header": "Set-Cookie","value": "lui=VjZnM1N0eGlYQnNZVlNjeTJHWjI0UT09;path=/;domain=.xxx.com;HttpOnly"}
        ],

         "bodyFile": "./respFiles/getUserInfo1.json"
        
      }
    }
```
与构造响应的规则结构与规则均一致，不多说了，但setHeaders、setBody都可为null.
另外针对json响应，如果bodyFile中的文件名不存在，这里会将服务器当前响应的json自动保存下来。

# 【五】响应文件（存放respFiles中）说明
即存放需要替换响应的文件，比如json、图片、html等。
比如，通过chrome获取的接口响应内容json，复制并保存下来(需带.json后缀)，在rules中配置好路径，就可以用了。内容随便改，响应随便变。
响应头中默认会设置一些常用文件的content-type，如果特殊的文件，需要自己配置setHeaders中content-type

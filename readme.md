# go_httpmock
     一个简单的使用goproxy作为代理的mock工具，支持http/https(添加x509.cer为根目录信任证书)
     这里是修改响应内容，前提是有服务端有响应
     可以修改rules.json中的设置响应头及相应内容
     rules.json改动不需要重启应用
     goproxy的性能不是很好，但作为mock工具来用差不多

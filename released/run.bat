@echo off
explorer .\respFiles | start http://127.0.0.1:8088 | goproxy_mock.exe
# 编译

```
go build boom.go
```

# 配置文件

```
globalheaders:
- key: Authorization
  value: Bearer token
scene:
  - name: 场景1
    urls:
    - url: http://127.0.0.1:8000/post/test
      method: POST
      data: test=123
      name: 测试post
      headers:
      - key: Content-Type
        value: application/x-www-form-urlencoded
    - url: https://www.baidu.com
      method: GET
      name: 获取百度首页
```

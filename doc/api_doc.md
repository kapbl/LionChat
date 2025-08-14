## dto
### Response Code
| 失败状态码，奇数 | 说明 |
| --- | --- |
| 1001 | json格式错误 |
| 1003 | 用户名或密码错误 |
| 1005 | 用户或者邮箱已存在 |
| 1009 | 数据库写入失败 |
| 1011 | 数据库查询失败 |
| 1013 | 数据库删除失败 |
| 1015 | 数据库更新失败 |
| 1017 | 生成token失败 |
### Success Code
| 成功状态码，偶数 | 说明 |
| --- | --- |
| 2000 | 注册成功 |
| 2002 | 登陆成功 |



### 
```go
// 基本的响应
type BaseResponse struct{
    // 请求ID, 由服务端生成
    RequestID string `json:"request_id"`
}
```
```go
// 登录响应
type LoginResponse struct{
    // 请求ID, 由服务端生成
    BaseResponse
    Code int32 `json:"code"`
    AccessToken string `json:"access_token"`
}
```
```go
// 登录请求
type LoginRequest struct{
    Account string `json:"account"`
    Password string `json:"password"`
}
```
```go
// 注册响应
type RegisterResponse struct{
    // 请求ID, 由服务端生成
    BaseResponse
    Code int32 `json:"code"`
    Msg string `json:"msg"`
}
```
```go
// 注册请求
type RegisterRequest struct{
    Email string `json:"email",validate:"required,email"`
    Username string `json:"username",validate:"required"`
    Nickname string `json:"nickname",validate:"required"`
    Password string `json:"password",validate:"required,min=6,max=12"`
}
```
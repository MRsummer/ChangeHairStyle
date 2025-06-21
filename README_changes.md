# 最近修改总结

## 1. 发型生成接口错误处理优化

### 修改文件：`pkg/handler/hair_style.go`

- 添加了429错误码的特殊处理
- 当火山引擎API返回429错误时，返回友好的提示信息："当前使用人数较多、请过5秒后尝试"
- 支持image_url和base64_image两种图片输入方式

## 2. 用户接口字段优化

### 修改文件：`pkg/model/user.go` 和 `pkg/handler/user.go`

#### 微信登录接口 (`/api/user/wx-login`)
- 添加了 `status` 字段，从环境变量 `USER_STATUS` 读取
- 将 `invite_code` 字段重命名为 `code`
- 将 `used_invite_code` 字段重命名为 `used_code`

#### 获取用户信息接口 (`/api/user/info/get`)
- 添加了 `status` 字段，从环境变量 `USER_STATUS` 读取
- 将 `invite_code` 字段重命名为 `code`
- 将 `used_invite_code` 字段重命名为 `used_code`

#### 使用邀请码接口 (`/api/user/invite-code/use`)
- 将请求参数中的 `invite_code` 字段重命名为 `code`

### 新增功能
- 添加了 `getStatusFromEnv()` 函数，用于从环境变量读取用户状态
- 环境变量 `USER_STATUS` 未设置或解析失败时，默认返回0

## 3. 测试工具

### 发型生成接口测试工具：`test_hair_style.go`

功能特性：
- 支持10个用户并行测试
- 使用base64图片数据
- 每个用户使用不同的发型提示词
- 预定义的用户ID列表
- 详细的统计信息和日志记录

使用方法：
```bash
# 编译
go build -o test_hair_style test_hair_style.go

# 运行（使用默认参数）
./test_hair_style

# 指定参数
./test_hair_style -url "https://your-server.com" -image "~/path/to/image.jpg" -verbose
```

### 完整用户场景测试工具：`test_user_scenario_advanced.go`

功能特性：
- 支持100个用户并行操作
- 模拟完整的用户操作流程
- 详细的统计信息
- 命令行参数配置

## 4. 环境变量配置

需要在环境变量中设置：
- `USER_STATUS`: 用户状态值（数字）

## 5. 接口响应示例

### 微信登录接口响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "test_user_123",
    "nickname": "用户昵称",
    "avatar_url": "头像URL",
    "coin": 60,
    "code": "INVITE123",
    "used_code": "USED456",
    "last_sign_in_date": "2024-01-15T10:30:00Z",
    "status": 1
  }
}
```

### 发型生成接口响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "image_url": "生成的图片URL",
    "record_id": 123
  }
}
```

### 429错误响应
```json
{
  "code": 500,
  "message": "当前使用人数较多、请过5秒后尝试"
}
```

### 使用邀请码接口

#### 请求示例
```json
{
  "user_id": "test_user_123",
  "code": "ABC123"
}
```

#### 响应示例
```json
{
  "code": 0,
  "message": "success"
}
``` 
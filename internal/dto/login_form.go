// ═════════════════════════════════════════════════════════════════════
// 数据传输对象（DTO）— 定义 API 请求/响应的数据格式
// 使用 binding 标签做参数校验，减少 handler 中的校验代码
// ═════════════════════════════════════════════════════════════════════
package dto

// LoginFormDTO 用户登录表单
// 支持两种方式：验证码登录（phone+code）或密码登录（phone+password）
type LoginFormDTO struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code"`     // 短信验证码
	Password string `json:"password"` // 密码（与 code 二选一）
}

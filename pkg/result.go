// ═════════════════════════════════════════════════════════════════════
// 统一 API 响应结构 — 所有 HTTP 接口统一返回格式
// 前端根据 success 字段判断成功/失败，errorMsg 展示错误信息
// 分页查询额外返回 total 字段
// ═════════════════════════════════════════════════════════════════════
package pkg

// Result 统一 JSON 响应，匹配前端 Vue3 项目的期望格式
// 相比标准 HTTP 状态码，业务层更倾向于通过 success+errorMsg 表达结果
type Result struct {
	Success  bool        `json:"success"`            // 业务是否成功
	ErrorMsg string      `json:"errorMsg"`           // 错误描述（成功时为空）
	Data     interface{} `json:"data"`               // 响应数据体
	Total    int64       `json:"total"`              // 分页总数（非分页查询为 0）
}

func OK() *Result {
	return &Result{Success: true}
}

func OKWithData(data interface{}) *Result {
	return &Result{Success: true, Data: data}
}

func OKWithDataTotal(data interface{}, total int64) *Result {
	return &Result{Success: true, Data: data, Total: total}
}

func Fail() *Result {
	return &Result{Success: false, ErrorMsg: "系统错误，请稍后重试!"}
}

func FailWithMsg(msg string) *Result {
	return &Result{Success: false, ErrorMsg: msg}
}

func FailWithError(err error) *Result {
	return &Result{Success: false, ErrorMsg: err.Error()}
}

func FailWithCode(err *BaseCodeError) *Result {
	return &Result{Success: false, ErrorMsg: err.Error()}
}

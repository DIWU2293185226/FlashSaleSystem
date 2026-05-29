// ═════════════════════════════════════════════════════════════════════
// Lua 脚本管理器 — 加载 + 缓存 SHA + EVALSHA 执行 + NOSCRIPT 回退
// 负责管理秒杀业务中用到的全部 Redis Lua 脚本
// 启动时通过 SCRIPT LOAD 将所有脚本加载到 Redis，缓存 SHA 摘要
// 执行时优先使用 EVALSHA（减少网络传输量），遇 NOSCRIPT 错误自动重载后重试
// ═════════════════════════════════════════════════════════════════════
package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LuaScript 表示一个已加载的 Lua 脚本，包含脚本名和 SHA 摘要
type LuaScript struct {
	Name string // 脚本名称（不含 .lua 后缀）
	SHA  string // SCRIPT LOAD 返回的 SHA 摘要
}

// LuaManager Lua 脚本管理器
// 维护脚本名 → LuaScript 的映射，提供统一的加载和执行接口
type LuaManager struct {
	redis   *RedisCache
	scripts map[string]*LuaScript
	dir     string // Lua 脚本文件目录
}

// NewLuaManager 创建 Lua 脚本管理器
// dir 参数指定 .lua 脚本文件的目录路径
func NewLuaManager(redis *RedisCache, dir string) *LuaManager {
	return &LuaManager{
		redis:   redis,
		scripts: make(map[string]*LuaScript),
		dir:     dir,
	}
}

// Load 从文件读取 .lua 脚本并通过 SCRIPT LOAD 加载到 Redis
// name 是脚本名（不含 .lua 后缀），对应的文件为 {dir}/{name}.lua
// 加载成功后缓存 SHA 摘要，后续执行使用 EVALSHA 代替 EVAL
func (m *LuaManager) Load(name string) error {
	scriptPath := filepath.Join(m.dir, name+".lua")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read lua script %s: %w", name, err)
	}

	ctx := context.Background()
	sha, err := m.redis.ScriptLoad(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to load lua script %s: %w", name, err)
	}

	m.scripts[name] = &LuaScript{Name: name, SHA: sha}
	return nil
}

// LoadAll 加载指定目录下所有 .lua 文件
// 遍历目录，对每个 .lua 文件调用 Load
func (m *LuaManager) LoadAll() error {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("failed to read lua directory %s: %w", m.dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".lua")
		if err := m.Load(name); err != nil {
			return err
		}
	}
	return nil
}

// Eval 使用 EVALSHA 执行已加载的 Lua 脚本
// 如果 Redis 返回 NOSCRIPT 错误（脚本在 Redis 中被 FLUSH 掉），
// 会自动重新加载脚本再重试，对上层透明
func (m *LuaManager) Eval(name string, keys []string, args ...interface{}) (interface{}, error) {
	script, ok := m.scripts[name]
	if !ok {
		return nil, fmt.Errorf("lua script %s not loaded", name)
	}

	ctx := context.Background()
	vals, err := m.redis.EvalSha(ctx, script.SHA, keys, args...)
	// NOSCRIPT 回退：如果 Redis 重启或 FLUSH 导致脚本丢失，自动重载
	if err != nil && strings.Contains(err.Error(), "NOSCRIPT") {
		if loadErr := m.Load(name); loadErr != nil {
			return nil, fmt.Errorf("failed to reload script %s: %w", name, loadErr)
		}
		script = m.scripts[name]
		return m.redis.EvalSha(ctx, script.SHA, keys, args...)
	}
	return vals, err
}

// EvalRaw 直接执行原始的 Lua 脚本字符串（不使用 SCRIPT LOAD/EVALSHA）
// 适用于一次性脚本或动态生成的脚本，不需要缓存 SHA
func (m *LuaManager) EvalRaw(script string, keys []string, args ...interface{}) (interface{}, error) {
	return m.redis.Eval(context.Background(), script, keys, args...)
}

// GetSHA 返回已加载脚本的 SHA 摘要，用于日志记录或调试
// 如果脚本未加载返回空字符串
func (m *LuaManager) GetSHA(name string) string {
	if s, ok := m.scripts[name]; ok {
		return s.SHA
	}
	return ""
}

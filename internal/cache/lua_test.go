package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLuaManager_NilRedis(t *testing.T) {
	lm := NewLuaManager(nil, "/nonexistent")
	assert.NotNil(t, lm)
	assert.Empty(t, lm.scripts)
}

func TestNewLuaManager_LoadInvalidDir(t *testing.T) {
	lm := NewLuaManager(nil, "/nonexistent")
	err := lm.Load("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read lua script")
}

func TestNewLuaManager_LoadAllInvalidDir(t *testing.T) {
	lm := NewLuaManager(nil, "/nonexistent")
	err := lm.LoadAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read lua directory")
}

func TestLuaManager_EvalNotLoaded(t *testing.T) {
	lm := NewLuaManager(nil, "/nonexistent")
	_, err := lm.Eval("not_loaded", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not loaded")
}

func TestLuaManager_GetSHA(t *testing.T) {
	lm := NewLuaManager(nil, "/nonexistent")
	assert.Empty(t, lm.GetSHA("anything"))
}

func TestNewLocalCache_DefaultSize(t *testing.T) {
	c := NewLocalCache(0)
	assert.NotNil(t, c)
	assert.Equal(t, 0, c.EntryCount())
}

func TestNewLocalCache_CustomSize(t *testing.T) {
	c := NewLocalCache(1024)
	assert.NotNil(t, c)
}

func TestLocalCache_SetGet(t *testing.T) {
	c := NewLocalCache(1024 * 10)

	err := c.Set("test:string", "hello", 60)
	assert.NoError(t, err)

	var val string
	err = c.Get("test:string", &val)
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestLocalCache_GetMiss(t *testing.T) {
	c := NewLocalCache(1024)
	var val string
	err := c.Get("nonexistent", &val)
	assert.Error(t, err)
}

func TestLocalCache_Del(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	_ = c.Set("test:del", "value", 60)

	assert.True(t, c.Del("test:del"))
	assert.False(t, c.Del("test:del")) // second delete should return false
}

func TestLocalCache_SetNull(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	err := c.SetNull("test:null")
	assert.NoError(t, err)
	assert.True(t, c.IsNull("test:null"))
}

func TestLocalCache_RawOperations(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	data := []byte{1, 2, 3, 4}
	err := c.SetRaw("test:raw", data, 60)
	assert.NoError(t, err)

	got, err := c.GetRaw("test:raw")
	assert.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestLocalCache_Clear(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	_ = c.Set("test:clear1", "v1", 60)
	_ = c.Set("test:clear2", "v2", 60)
	c.Clear()
	assert.Equal(t, 0, c.EntryCount())
}

func TestLocalCache_HitRate(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	_ = c.Set("test:hit", "value", 60)
	// Miss first
	var v string
	_ = c.Get("test:miss", &v)
	// Then hit
	_ = c.Get("test:hit", &v)
	assert.True(t, c.HitRate() > 0)
}

func TestLocalCache_TTL(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	_ = c.Set("test:ttl", "value", 60)
	assert.Equal(t, 0, c.TTL("test:ttl")) // found
	assert.Equal(t, -1, c.TTL("nonexistent")) // not found
}

func TestLocalCache_Expire(t *testing.T) {
	c := NewLocalCache(1024 * 10)
	_ = c.Set("test:exp", "value", 60)
	assert.True(t, c.Expire("test:exp", 120))
	assert.False(t, c.Expire("nonexistent", 120))
}

func TestNewMultiLevelCache(t *testing.T) {
	mc := NewMultiLevelCache(0, nil)
	assert.NotNil(t, mc)
	assert.NotNil(t, mc.local)
	assert.Nil(t, mc.redis)
}

func TestNewMultiLevelCacheWithLocal(t *testing.T) {
	local := NewLocalCache(1024)
	mc := NewMultiLevelCacheWithLocal(local, nil)
	assert.NotNil(t, mc)
	assert.Equal(t, local, mc.local)
}

func TestScopedCache_LocalOnly(t *testing.T) {
	local := NewLocalCache(1024 * 10)
	mc := NewMultiLevelCacheWithLocal(local, nil)
	scoped := mc.WithLevel(CacheLevelLocal)

	err := scoped.Set("test:scoped", "scoped-value", 60)
	assert.NoError(t, err)

	var val string
	err = scoped.Get("test:scoped", &val)
	assert.NoError(t, err)
	assert.Equal(t, "scoped-value", val)
}

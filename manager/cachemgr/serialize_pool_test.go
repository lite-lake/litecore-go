package cachemgr

import (
	"bytes"
	"testing"
)

// TestSerializeWithPool_DoesNotAliasPooledBuffer 回归：序列化结果不得引用对象池 buffer。
// 并发 SET 场景下，buffer 归还池后会被其他 goroutine 复用；若直接返回 buf.Bytes()，
// 已序列化的数据会被改写，导致把 A 的内容写进 B 的 key。
func TestSerializeWithPool_DoesNotAliasPooledBuffer(t *testing.T) {
	data, err := serializeWithPool("first-value")
	if err != nil {
		t.Fatalf("serializeWithPool: %v", err)
	}

	// 模拟另一个 goroutine 取到同一个池化 buffer 并写入不同内容
	buf := gobPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteString("second-value-corrupted")
	gobPool.Put(buf)

	var got string
	if err := deserializeWithPool(data, &got); err != nil {
		t.Fatalf("序列化结果被池化 buffer 复用污染: %v", err)
	}
	if got != "first-value" {
		t.Fatalf("序列化结果应为 first-value，实际 %q", got)
	}
}

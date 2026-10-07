//go:build linux

package libgopeed

import (
	"log"
	"syscall"
)

// 2026-10-05 高速核心性能修复：
// 多任务同时开高线程（例如每个任务 512 连接）时，Android 进程默认的 fd 软上限
// （常见 1024）会先被撞到下（EMFILE），表现为「连接建不起来 / 下载卡死 / 速度骤降」，
// 而不是带宽或内核调度问题。启动时把软上限抬到硬上限（并尽量抬硬上限到 1<<20）。
func init() { raiseFdLimit() }

func raiseFdLimit() {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return
	}
	cur, max := lim.Cur, lim.Max

	const want = 1 << 20
	if max < want {
		try := syscall.Rlimit{Cur: want, Max: want}
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &try); err == nil {
			cur, max = try.Cur, try.Max
		}
	}
	if cur < max {
		lim.Cur, lim.Max = max, max
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &lim); err == nil {
			cur = max
		}
	}
	log.Printf("[gopeed] RLIMIT_NOFILE cur=%d max=%d (多任务高线程需求)", cur, max)
}

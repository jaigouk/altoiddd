//go:build windows

package infrastructure

// noFollow is a no-op on Windows — the O_NOFOLLOW flag is POSIX-specific
// and Windows symlink semantics differ enough that the plan-phase Lstat
// sweep in WriteScaffold is the binding defence on this platform. Defined
// here so the cross-platform call site can OR in noFollow unconditionally.
const noFollow = 0

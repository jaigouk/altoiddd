//go:build unix

package infrastructure

import "syscall"

// noFollow is OR'd into OpenFile flags so the kernel refuses to traverse a
// symlink at the final path component. On POSIX systems syscall.O_NOFOLLOW
// causes open(2) to fail with ELOOP if the target is a symlink — a defence
// against an attacker pre-planting a symlink under <targetDir>/alto-scaffold/ to
// trick --force into clobbering a sensitive file elsewhere. The plan-phase
// Lstat sweep in WriteScaffold is the primary defence; this flag is the
// kernel-level backstop.
const noFollow = syscall.O_NOFOLLOW

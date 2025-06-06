// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

package sha1

import (
	"crypto/internal/impl"
	"internal/cpu"
)

//go:noescape
func blockAVX2(dig *digest, p []byte)

//go:noescape
func blockSHANI(dig *digest, p []byte)

var useAVX2 = cpu.X86.HasAVX && cpu.X86.HasAVX2 && cpu.X86.HasBMI1 && cpu.X86.HasBMI2
var useSHANI = cpu.X86.HasAVX && cpu.X86.HasSHA && cpu.X86.HasSSE41 && cpu.X86.HasSSSE3

func init() {
	impl.Register("sha1", "AVX2", &useAVX2)
	impl.Register("sha1", "SHA-NI", &useSHANI)
}

func block(dig *digest, p []byte) {
	if useSHANI {
		blockSHANI(dig, p)
	} else if useAVX2 && len(p) >= 256 {
		// blockAVX2 calculates sha1 for 2 block per iteration and also
		// interleaves precalculation for next block. So it may read up-to 192
		// bytes past end of p. We could add checks inside blockAVX2, but this
		// would just turn it into a copy of the old pre-AVX2 amd64 SHA1
		// assembly implementation, so just call blockGeneric instead.
		safeLen := len(p) - 128
		if safeLen%128 != 0 {
			safeLen -= 64
		}
		blockAVX2(dig, p[:safeLen])
		blockGeneric(dig, p[safeLen:])
	} else {
		blockGeneric(dig, p)
	}
}

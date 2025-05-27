/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package types

const (
	// byte codec versions
	codec_RawDynoBuffer = byte(0x00) + iota
	codec_RDB_1         // + row system fields mask
	codec_RDB_2         // + CUD row emptied fields

	// !do not forget to actualize last codec version!
	codec_LastVersion = codec_RDB_2
)

// system fields mask values
const (
	sfm_ID        = uint16(1 << 0)
	sfm_ParentID  = uint16(1 << 1)
	sfm_Container = uint16(1 << 2)
	sfm_IsActive  = uint16(1 << 3)
)

// maskString is character to mask values in string cell, used for obfuscate unlogged command arguments data
const maskString = "*"

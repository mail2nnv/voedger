/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package rows

const (
	// byte codec versions
	codec_RawDynoBuffer = byte(0x00) + iota
	codec_RDB_1         // + row system fields mask
	codec_RDB_2         // + CUD row emptied fields

	// !do not forget to actualize last codec version!
	codec_LastVersion = codec_RDB_2
)

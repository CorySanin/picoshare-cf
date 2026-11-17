package handlers

import "codefloe.com/mtlynch/picoshare/space"

// SpaceCheckFunc measures PicoShare storage usage.
type SpaceCheckFunc func() (space.Usage, error)

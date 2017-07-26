package tarutil

const (
	whiteoutPrefix          = ".wh."
	whiteoutMetaPrefix      = whiteoutPrefix + whiteoutPrefix
	whiteoutLinkDir         = whiteoutMetaPrefix + "plnk"
	whiteoutOpaqueDir       = whiteoutMetaPrefix + ".opq"
	overlayOpaqueXattr      = "trusted.overlay.opaque"
	overlayOpaqueXattrValue = "y"
)

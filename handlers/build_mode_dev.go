//go:build dev

package handlers

// Development builds must always serve the latest frontend source files.
const staticCacheControl = "no-store"

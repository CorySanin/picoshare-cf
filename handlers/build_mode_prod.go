//go:build !dev

package handlers

// Production builds may cache embedded assets, but bound the lifetime so that
// clients eventually recover if a deployment reuses a process or URL.
const staticCacheControl = "public, max-age=3600"

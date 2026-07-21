package version

// Version is Ratatoskr's release version. Overridable at build time with:
//
//	go build -ldflags "-X github.com/Oriotic/Ratatoskr/internal/version.Version=1.2.3"
var Version = "1.0-dev"

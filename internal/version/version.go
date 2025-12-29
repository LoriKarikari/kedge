package version

var (
	version = "dev"
	commit  = "unknown"
)

func Version() string { return version }
func Commit() string  { return commit }

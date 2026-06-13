package buildinfo

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

func Current() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
}

package version

var version = "?"
var buildTime = "?"
var commit = "?"

type Version struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"` // Time when application was build.
	Commit    string `json:"commit"`     // Git commit hash.
}

func GetVersion() *Version {
	return &Version{
		Version:   version,
		BuildTime: buildTime,
		Commit:    commit,
	}
}

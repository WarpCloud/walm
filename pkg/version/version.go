package version

import "fmt"

var (
	Version   = "0.0.1"
	GitSha1Version = ""
	BuildDate	 = "0000-00-00"
)

func PrintVersionInfo() {
	fmt.Println("Release Version: ", Version)
	fmt.Println("Git Commit Hash: ", GitSha1Version)
	fmt.Println("Build Date: ", BuildDate)
}

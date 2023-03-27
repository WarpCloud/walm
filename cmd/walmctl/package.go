package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var longPackageHelp = `
This command packages a chart into a versioned chart archive file. If a path
is given, this will look at that path for a chart (which must contain a
Chart.yaml file) and then package that directory.

If no path is given, this will look in the present working directory for a
Chart.yaml file, and (if found) build the current directory into a chart.

Versioned chart archives are used by Helm package repositories.
`

type packageOptions struct {
	destination string
	chartPath   string
}

func newPackageCmd(out io.Writer) *cobra.Command {
	pack := &packageOptions{chartPath: "."}

	cmd := &cobra.Command{
		Use:   "package [CHART_PATH] [...]",
		Short: "package a chart directory into a chart archive",
		Long:  longPackageHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pack.run()
		},
	}
	cmd.PersistentFlags().StringVarP(&pack.chartPath, "chartPath", "c", ".", "transwarp chart path")
	cmd.PersistentFlags().StringVarP(&pack.destination, "destination", "d", ".", "location to write the chart")
	cmd.MarkPersistentFlagRequired("chartPath")
	return cmd
}

func (o *packageOptions) run() error {
	chartPath, err := filepath.Abs(o.chartPath)
	if err != nil {
		return err
	}

	//copy charts to tmpdir
	tmpdir, err := createTempDir()
	if err != nil {
		return err
	}
	err = copyDir(chartPath, tmpdir)
	if err != nil {
		return err
	}

	// cp applib to tmpdir
	applibPath := filepath.Join(chartPath, "../../applib")
	destPath := filepath.Join(tmpdir, "applib")
	err = copyDir(applibPath, destPath)
	if err != nil {
		return err
	}

	path := tmpdir

	// append ci/ dir to .helmIgnore
	ignoreFiles := []string{"ci/", "applib/ksonnet-lib/"}
	err = appendHelmIgnoreFile(path, ignoreFiles)
	if err != nil {
		return err
	}

	ch, err := loader.LoadDir(path)
	if err != nil {
		return err
	}

	validChartType := isValidChartType(ch.Metadata.Type)
	if !validChartType {
		return errors.New("not a ValidChartType")
	}

	var dest string
	if o.destination == "." {
		// Save to the current working directory.
		dest, err = os.Getwd()
		if err != nil {
			return err
		}
	} else {
		// Otherwise save to set destination
		dest = o.destination
	}

	name, err := chartutil.Save(ch, dest)
	if err != nil {
		return errors.Wrap(err, "failed to save")
	}
	fmt.Fprintf(os.Stdout, "Successfully packaged chart and saved it to: %s\n", name)

	_, err = os.Stat(tmpdir)
	if err == nil {
		err = os.RemoveAll(tmpdir)
		if err != nil {
			return err
		}
	}

	return nil
}

func appendHelmIgnoreFile(path string, ignoreFiles []string) error {
	ifile := filepath.Join(path, ".helmignore")
	f, err := os.OpenFile(ifile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, tmpFile := range ignoreFiles {
		f.WriteString(tmpFile)
		f.WriteString("\n")
	}

	return nil
}

func copyDir(srcPath string, destPath string) error {
	if srcInfo, err := os.Stat(srcPath); err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		if !srcInfo.IsDir() {
			e := errors.New(fmt.Sprintf("%s is not correct dir \n", srcInfo))
			return e
		}
	}

	if destInfo, err := os.Stat(destPath); err != nil {
		err := os.Mkdir(destPath, os.ModePerm)
		if err != nil {
			return err
		}

	} else {
		if !destInfo.IsDir() {
			e := errors.New(fmt.Sprintf("%s is not correct dir \n", destInfo))
			return e
		}
	}

	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			destNewPath := strings.Replace(path, srcPath, destPath, -1)
			copyFile(path, destNewPath)
		}
		return nil

	})

	if err != nil {
		fmt.Printf(err.Error())
	}
	return err
}

func copyFile(src, dest string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()

	destSplitPathDirs := strings.Split(dest, "/")

	destSplitPath := ""
	for index, dir := range destSplitPathDirs {
		if index < len(destSplitPathDirs)-1 {
			destSplitPath = destSplitPath + dir + "/"
			b, _ := pathExists(destSplitPath)
			if b == false {
				err := os.Mkdir(destSplitPath, os.ModePerm)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	dstFile, err := os.Create(dest)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createTempDir() (string, error) {
	dir, err := ioutil.TempDir("/tmp/", "tmp")
	if err != nil {
		fmt.Println("create tmpdir fail")
		return "", err
	}

	return dir, nil
}

func isValidChartType(in string) bool {
	switch in {
	case "", "application", "library":
		return true
	}
	return false
}

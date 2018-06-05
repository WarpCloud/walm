package cluster

type Deps struct {
	Appname string
	Deps    []string
}

type DepList []Deps

type Graph map[int][]string

//param: name of chart
func GetDeps(name string) ([]string, error) {
	return []string{}, nil
}

//param: name of product
func MakeGraph(name string) error {
	//get applications list with product name
	var apps []string

	//get all deps of applications
	var dl DepList
	var err error
	for _, app := range apps {
		dep := Deps{Appname: app}
		if dep.Deps, err = GetDeps(app); err != nil {
			return err
		}
		dl = append(dl, dep)
	}

	//get graph to install
	for _, dep := range dl {

	}
}

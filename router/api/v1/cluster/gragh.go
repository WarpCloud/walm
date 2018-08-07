package cluster

import (
	"fmt"
	"sync/atomic"
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
)

var id uint64

type DepInterface interface {
	getDeps(leaf *Leaf) []Leaf
}

type DepInst struct{}

var depInst DepInterface

func init() {
	depInst = &DepInst{}
}

type Leaf struct {
	app   *Application
	color bool
	level int
	next  *Leaf
}

type Application struct {
	inst     *release.ReleaseRequest
	Depend   map[string]*release.ReleaseRequest
	Bedepend []*Application
}

func getGraghForInstance(name, namespace string, releaeMap map[string]string, inst *release.ReleaseRequest) (err error, instList []release.ReleaseRequest) {
	a_left := []Leaf{}
	leaf_1 := Leaf{
		app: &Application{
			inst:     inst,
			Depend:   map[string]*release.ReleaseRequest{},
			Bedepend: []*Application{},
		},
		color: false,
		level: 0,
		next:  nil,
	}
	a_left = append(a_left, leaf_1)
	getDepArray(name, namespace, &leaf_1, &a_left)

	root := &Leaf{level: -1}
color:
	for index, leaf_3 := range a_left {
		if leaf_3.color {
			continue
		}
		curr := leaf_3
		for index_1, leaf_4 := range a_left[index+1:] {
			if isSameLeaf(&curr, &leaf_4) {
				if curr.level < leaf_4.level {
					curr.color = true
					for _, cab := range curr.app.Bedepend {
						cab.Depend[curr.app.inst.ChartName] = leaf_4.app.inst
					}
					curr = leaf_4
				} else {
					a_left[index+1+index_1].color = true
					for _, cab := range leaf_4.app.Bedepend {
						cab.Depend[leaf_4.app.inst.ChartName] = curr.app.inst
					}
				}
			}
		}
		if root.level >= curr.level {
			curr.next = root
			root = &curr
		} else {
			p1, p2 := root, root.next
			for p2 != nil {
				if p2.level > curr.level {
					curr.next = p2
					p1.next = &curr
					continue color
				}
				p1 = p2
				p2 = p2.next
			}
			p1.next = &curr
			curr.next = nil
		}
	}

	list := []Leaf{}
	for p := root.next; p != nil; p = p.next {
		list = append(list, *p)
	}

	for index, leaf := range list {
		for release, inst := range leaf.app.Depend {
			if nameNew, ok := releaeMap[release]; ok {
				leaf.app.Depend[release].Name = nameNew
				func() {
					for index_1, leaf := range list[index:] {
						if leaf.app.inst.Name == inst.Name {
							list[index_1].color = true
							return
						}
					}
				}()
			}
		}
	}
	for _, leaf := range list {
		if !leaf.color {
			instList = append(instList, *expandDep(&leaf))
		}
	}
	return
}

func getGragh(name, namespace string, cluster *Cluster) (err error, instList []release.ReleaseRequest) {
	a_left := []Leaf{}

	for _, app := range cluster.Apps {
		appcopy := app
		leaf_1 := Leaf{
			app: &Application{
				inst:     &appcopy,
				Depend:   map[string]*release.ReleaseRequest{},
				Bedepend: []*Application{},
			},
			color: false,
			level: 0,
			next:  nil,
		}
		a_left = append(a_left, leaf_1)
		getDepArray(name, namespace, &leaf_1, &a_left)
	}

	root := &Leaf{level: -1}

color:
	for index, leaf_3 := range a_left {
		if leaf_3.color {
			continue
		}
		curr := leaf_3
		for index_1, leaf_4 := range a_left[index+1:] {
			if isSameLeaf(&curr, &leaf_4) {
				if curr.level < leaf_4.level {
					curr.color = true
					for _, cab := range curr.app.Bedepend {
						cab.Depend[curr.app.inst.ChartName] = leaf_4.app.inst
					}
					curr = leaf_4
				} else {
					a_left[index+1+index_1].color = true
					for _, cab := range leaf_4.app.Bedepend {
						cab.Depend[leaf_4.app.inst.ChartName] = curr.app.inst
					}
				}
			}
		}
		if root.level <= curr.level {
			curr.next = root
			root = &curr
		} else {
			p1, p2 := root, root.next
			for p2 != nil {
				if p2.level < curr.level {
					curr.next = p2
					p1.next = &curr
					continue color
				}
				p1 = p2
				p2 = p2.next
			}
			p1.next = &curr
			curr.next = nil
		}
	}

	for p := root; p.level >= 0; p = p.next {
		instList = append(instList, *expandDep(p))
	}

	return nil, instList
}

//godoc
//expand depend of instance
func expandDep(leaf *Leaf) *release.ReleaseRequest {

	for name, dep := range leaf.app.Depend {
		leaf.app.inst.Dependencies[name] = dep.Name
	}
	return leaf.app.inst
}

//godoc
//get all the depences from leaf, will go through to the end
func getDepArray(name, namespace string, leaf *Leaf, a_leaf *[]Leaf) []Leaf {
	for _, leaf_1 := range depInst.getDeps(leaf) {
		leaf_1.level = leaf.level + 1

		if len(leaf_1.app.inst.Name) == 0 {
			leaf_1.app.inst.Name = fmt.Sprintf("%s-%s-%s-%010d", namespace, name, leaf_1.app.inst.ChartName, atomic.AddUint64(&id, 1))
			leaf_1.app.inst.Namespace = namespace
		}
		//consider there is nome chart (same chart type and not same chart version in the same cluster)
		leaf.app.Depend[leaf_1.app.inst.ChartName] = leaf_1.app.inst
		leaf_1.app.Bedepend = append(leaf_1.app.Bedepend, leaf.app)
		*a_leaf = append(*a_leaf, leaf_1)
		getDepArray(name, namespace, &leaf_1, a_leaf)
	}
	return *a_leaf
}

//godoc
//get depences from leaf, only one level
func (di *DepInst) getDeps(leaf *Leaf) []Leaf {
	var leafArray []Leaf
	chartName, chartVersion := leaf.app.inst.ChartName, leaf.app.inst.ChartVersion
	if nameList, versionList, err := helm.GetDependencies(chartName, chartVersion); err != nil {
		return []Leaf{}
	} else {
		for index, name := range nameList {
			version := versionList[index]
			leafArray = append(leafArray, Leaf{
				app: &Application{
					inst: &release.ReleaseRequest{
						ChartName:    name,
						ChartVersion: version,
					},
				},
			})
		}
		return leafArray
	}

}

//godoc
//compare is the same leaf ,if color is true skip it
func isSameLeaf(a, b *Leaf) bool {
	if b.color {
		return false
	}
	if a.app.inst.ChartName == b.app.inst.ChartName && a.app.inst.Name == b.app.inst.Name && a.app.inst.ChartVersion == b.app.inst.ChartVersion {
		return true
	} else {
		return false
	}
}

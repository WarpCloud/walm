package cluster

import (
	"strconv"
	"strings"
	"walm/router/api/v1/instance"
)

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
	inst      *instance.Application
	Depend    map[string]*instance.Application
	Bedepend  []*Application
	ClusterId int
}

func getGraghForInstance(clusterid int, releaeMap map[string]string, inst *instance.Application) (err error, instList []instance.Application) {
	a_left := []Leaf{}
	leaf_1 := Leaf{
		app: &Application{
			inst:      inst,
			ClusterId: clusterid,
			Depend:    map[string]*instance.Application{},
			Bedepend:  []*Application{},
		},
		color: false,
		level: 0,
		next:  nil,
	}
	a_left = append(a_left, leaf_1)
	getDepArray(&leaf_1, &a_left)

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
						cab.Depend[curr.app.inst.Chart] = leaf_4.app.inst
					}
					curr = leaf_4
				} else {
					a_left[index+1+index_1].color = true
					for _, cab := range leaf_4.app.Bedepend {
						cab.Depend[leaf_4.app.inst.Chart] = curr.app.inst
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

func getGragh(clusterid int, name, namespace string, cluster *Cluster) (err error, instList []instance.Application) {
	a_left := []Leaf{}

	for _, app := range cluster.Apps {
		appcopy := app
		leaf_1 := Leaf{
			app: &Application{
				inst:      &appcopy,
				ClusterId: clusterid,
				Depend:    map[string]*instance.Application{},
				Bedepend:  []*Application{},
			},
			color: false,
			level: 0,
			next:  nil,
		}
		a_left = append(a_left, leaf_1)
		getDepArray(&leaf_1, &a_left)
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
						cab.Depend[curr.app.inst.Chart] = leaf_4.app.inst
					}
					curr = leaf_4
				} else {
					a_left[index+1+index_1].color = true
					for _, cab := range leaf_4.app.Bedepend {
						cab.Depend[leaf_4.app.inst.Chart] = curr.app.inst
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
func expandDep(leaf *Leaf) *instance.Application {

	for name, dep := range leaf.app.Depend {
		chartname := strings.Split(name, ":")[0]
		leaf.app.inst.Links[chartname] = dep.Name
	}
	return leaf.app.inst
}

//godoc
//get all the depences from leaf, will go through to the end
func getDepArray(leaf *Leaf, a_leaf *[]Leaf) []Leaf {
	for _, leaf_1 := range depInst.getDeps(leaf) {
		leaf_1.level = leaf.level + 1
		if len(leaf_1.app.inst.Name) == 0 {
			leaf_1.app.inst.Name = leaf_1.app.inst.Chart + "_" + strconv.Itoa(leaf.app.ClusterId)
		}
		leaf.app.Depend[leaf_1.app.inst.Chart] = leaf_1.app.inst
		leaf_1.app.Bedepend = append(leaf_1.app.Bedepend, leaf.app)
		*a_leaf = append(*a_leaf, leaf_1)
		getDepArray(&leaf_1, a_leaf)
	}
	return *a_leaf
}

//godoc
//get depences from leaf, only one level
func (di *DepInst) getDeps(leaf *Leaf) []Leaf {
	return []Leaf{}
}

//godoc
//compare is the same leaf ,if color is true skip it
func isSameLeaf(a, b *Leaf) bool {
	if b.color {
		return false
	}
	if a.app.inst.Chart == b.app.inst.Chart && a.app.inst.Name == b.app.inst.Name {
		return true
	} else {
		return false
	}
}

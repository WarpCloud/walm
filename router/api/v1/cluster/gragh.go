package cluster

import (
	"strconv"
	"strings"
	"walm/router/api/v1/instance"
)

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

func getGragh(clusterid int, name, namespace string, cluster *Cluster) (err error, instList []instance.Application) {
	a_left := []*Leaf{}
	for _, app := range cluster.Apps {
		leaf_1 := Leaf{
			app:   &Application{inst: &app, ClusterId: clusterid},
			color: false,
			level: 0,
			next:  nil,
		}
		a_left = append(a_left, &leaf_1)
		getDepArray(&leaf_1, &a_left)
	}

	root := &Leaf{level: -1}

color:
	for index, leaf_3 := range a_left {
		if leaf_3.color {
			continue
		}
		curr := leaf_3
		for _, leaf_4 := range a_left[index:] {
			if isSameLeaf(curr, leaf_4) {
				if curr.level < leaf_4.level {
					curr.color = true
					for _, cab := range curr.app.Bedepend {
						cab.Depend[curr.app.inst.Chart] = leaf_4.app.inst
					}
					curr = leaf_4
				} else {
					leaf_4.color = true
					for _, cab := range leaf_4.app.Bedepend {
						cab.Depend[leaf_4.app.inst.Chart] = curr.app.inst
					}
				}
			}
		}
		if root.level <= curr.level {
			curr.next = root
			root = curr
		} else {
			p1, p2 := root, root.next
			for p2 != nil {
				if p2.level < curr.level {
					curr.next = p2
					p1.next = curr
					continue color
				}
				p1 = p2
				p2 = p2.next
			}
			p1.next = curr
			curr.next = nil
		}
	}

	for p := root; p != nil; p = p.next {
		instList = append(instList, *expandDep(p))
	}

	return nil, instList
}

//godoc
//expand depend of instance
func expandDep(leaf *Leaf) *instance.Application {

	for _, dep := range leaf.app.Depend {
		chartname := strings.Split(dep.Chart, ",")[0]
		leaf.app.inst.Links[chartname] = dep.Name
	}
	return leaf.app.inst
}

//godoc
//get all the depences from leaf, will go through to the end
func getDepArray(leaf *Leaf, a_leaf *[]*Leaf) []*Leaf {
	for _, leaf_1 := range getDeps(leaf) {
		leaf_1.level = leaf.level + 1
		leaf_1.app.inst.Name = leaf_1.app.inst.Chart + strconv.Itoa(leaf.app.ClusterId)
		leaf.app.Depend[leaf_1.app.inst.Chart] = leaf_1.app.inst
		leaf_1.app.Bedepend = append(leaf_1.app.Bedepend, leaf.app)
		*a_leaf = append(*a_leaf, &leaf_1)
		getDepArray(&leaf_1, a_leaf)
	}
	return []*Leaf{}
}

//godoc
//get depences from leaf, only one level
func getDeps(leaf *Leaf) []Leaf {
	return []Leaf{}
}

//godoc
//compare is the same leaf ,if color is true skip it
func isSameLeaf(a, b *Leaf) bool {
	if b.color {
		return true
	}
	if a.app.inst.Chart == b.app.inst.Chart && a.app.inst.Name == a.app.inst.Name {
		return true
	} else {
		return false
	}
}

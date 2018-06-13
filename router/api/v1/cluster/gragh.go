package cluster

import (
	"walm/router/api/v1/instance"
)

type Leaf struct {
	app   *instance.Application
	color bool
	level int
	left  *Leaf
	right *Leaf
}

func getGragh(name, namespace string, cluster *Cluster) (err error, instList []instance.Application) {
	a_left := []*Leaf{}
	for _, app := range cluster.Apps {
		leaf_1 := Leaf{
			app:   &app,
			color: false,
			level: 0,
			left:  nil,
			right: nil,
		}
		a_left = append(a_left, &leaf_1)
		getDepArray(&leaf_1, &a_left)
	}
	root := a_left[0]
	for index, leaf_3 := range a_left {
		if leaf_3.color {
			continue
		}
		curr := leaf_3
		for _, leaf_4 := range a_left[index:] {
			if isSameLeaf(curr, leaf_4) {
				if curr.level < leaf_4.level {
					curr.color = true
					curr = leaf_4
				} else {
					leaf_4.color = true
				}
			}
		}
		makeTree(root, curr)
	}
	printTree(root, &instList)
	return nil, instList
}

func makeTree(root *Leaf, leaf *Leaf) {
	if root == nil {
		root = leaf
	} else if root.level > leaf.level {
		leaf.left = root
		root = leaf
	} else if root.level == leaf.level {
		if root.right == nil {
			root.right = leaf
		} else {
			leaf.left = root
			root = leaf
		}
	} else {
		makeTree(root.left, leaf)
	}
}

//godoc
//print the whole tree
func printTree(root *Leaf, instList *[]instance.Application) {
	if root == nil {
		return
	} else if root.left != nil {
		printTree(root, instList)
	} else if root.right != nil {
		*instList = append(*instList, *root.right.app)
	}
	*instList = append(*instList, *root.app)
}

//godoc
//get all the depences from leaf, will go through to the end
func getDepArray(leaf *Leaf, a_leaf *[]*Leaf) []*Leaf {
	for _, leaf_1 := range getDeps(leaf) {
		leaf_1.level = leaf.level + 1
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
	if a.app.Chart == b.app.Chart && a.app.Name == a.app.Name {
		return true
	} else {
		return false
	}
}

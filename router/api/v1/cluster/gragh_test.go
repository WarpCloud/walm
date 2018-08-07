package cluster

import (
	"testing"

	"gopkg.in/check.v1"

	"walm/pkg/release"
)

func Test(t *testing.T) { check.TestingT(t) }

type clusterSuite struct{}

var _ = check.Suite(&clusterSuite{})

func (cs *clusterSuite) Test_expandDep(c *check.C) {
	leaf := &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test", Dependencies: map[string]string{}},
			Depend: map[string]*release.ReleaseRequest{
				"test_chart_1:1.0.0": &release.ReleaseRequest{Name: "test_1"},
				"test_chart_2":       &release.ReleaseRequest{Name: "test_2"},
			},
		},
	}
	app := expandDep(leaf)
	c.Assert(app.Dependencies["test_chart_1:1.0.0"], check.Equals, "test_1")
	c.Assert(app.Dependencies["test_chart_2"], check.Equals, "test_2")
}

func (cs *clusterSuite) Test_isSameLeaf(c *check.C) {
	a := &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0"},
		},
	}
	b := &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0"},
		},
		color: true,
	}

	re := isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

	b = &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, true)

	b = &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test_2", ChartName: "test_chart_1", ChartVersion: "1.0.0"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

	b = &Leaf{
		app: &Application{
			inst: &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.1"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

}

type mockDepInst struct{}

func (mdi *mockDepInst) getDeps(leaf *Leaf) []Leaf {
	if leaf.app.inst.ChartName == "test_chart_1" {
		return []Leaf{
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_1_1", ChartName: "test_chart_1_1", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_1_2", ChartName: "test_chart_1_2", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
		}
	}

	if leaf.app.inst.ChartName == "test_chart_2" {
		return []Leaf{
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_2_1", ChartName: "test_chart_2_1", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_1_2", ChartName: "test_chart_1_2", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
		}
	}
	if leaf.app.inst.ChartName == "test_chart_3_1" {
		return []Leaf{
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
			Leaf{
				app: &Application{
					inst:     &release.ReleaseRequest{Name: "test_2", ChartName: "test_chart_2", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
					Depend:   map[string]*release.ReleaseRequest{},
					Bedepend: []*Application{},
				},
			},
		}
	}
	return []Leaf{}
}

func (cs *clusterSuite) Test_getDepArray(c *check.C) {
	back := depInst
	depInst = &mockDepInst{}
	defer func() {
		depInst = back
	}()
	a_leaf := &[]Leaf{}
	leaf := &Leaf{
		app: &Application{
			inst:     &release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
			Depend:   map[string]*release.ReleaseRequest{},
			Bedepend: []*Application{},
		},
		color: false,
	}

	*a_leaf = append(*a_leaf, *leaf)

	depArray := getDepArray("test", "test", leaf, a_leaf)

	c.Assert(depArray, check.HasLen, 3)
	c.Assert(depArray[0].app.inst.Name, check.Equals, "test_1")
	c.Assert(depArray[1].app.inst.ChartName, check.Equals, "test_chart_1_1")
	c.Assert(depArray[1].app.inst.ChartVersion, check.Equals, "1.0.0")
	c.Assert(depArray[2].app.inst.ChartName, check.Equals, "test_chart_1_2")
	c.Assert(depArray[2].app.inst.ChartVersion, check.Equals, "1.0.0")
	c.Assert(leaf.app.Depend["test_chart_1_1"].Name, check.Equals, "test_1_1")
	c.Assert(leaf.app.Depend["test_chart_1_2"].Name, check.Equals, "test_1_2")
	c.Assert(depArray[1].app.Bedepend[0].inst.Name, check.Equals, "test_1")
	c.Assert(depArray[2].app.Bedepend[0].inst.Name, check.Equals, "test_1")
}

func (cu *clusterSuite) Test_getGragh(c *check.C) {
	back := depInst
	depInst = &mockDepInst{}
	defer func() {
		depInst = back
	}()

	cluster := &Cluster{
		Apps: []release.ReleaseRequest{
			release.ReleaseRequest{Name: "test_1", ChartName: "test_chart_1", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
			release.ReleaseRequest{Name: "test_2", ChartName: "test_chart_2", ChartVersion: "1.0.0", Dependencies: map[string]string{}},
		},
	}
	err, instList := getGragh("test", "test", cluster)

	c.Assert(err, check.IsNil)
	c.Assert(instList, check.HasLen, 5)
	c.Assert(instList[0].Name, check.Equals, "test_2_1")
	c.Assert(instList[1].ChartName, check.Equals, "test_chart_1_2")
	c.Assert(instList[1].ChartVersion, check.Equals, "1.0.0")
	c.Assert(instList[2].ChartName, check.Equals, "test_chart_1_1")
	c.Assert(instList[2].ChartVersion, check.Equals, "1.0.0")
	c.Assert(instList[3].Name, check.Equals, "test_1")
	c.Assert(instList[4].Name, check.Equals, "test_2")

}

func (cu *clusterSuite) Test_getGraghForInstance(c *check.C) {
	back := depInst
	depInst = &mockDepInst{}
	defer func() {
		depInst = back
	}()
	releaeMap := map[string]string{
		"test_chart_1": "test_1",
		"test_chart_2": "test_2"}
	inst := &release.ReleaseRequest{
		Name:         "test_3_1",
		ChartName:    "test_chart_3_1",
		ChartVersion: "1.0.0",
		Dependencies: map[string]string{},
	}
	err, instList := getGraghForInstance("test", "test", releaeMap, inst)
	c.Assert(err, check.IsNil)
	c.Assert(instList, check.HasLen, 4)
	c.Assert(instList[0].Name, check.Equals, "test_3_1")
	c.Assert(instList[0].Dependencies["test_chart_2"], check.Equals, "test_2")

}

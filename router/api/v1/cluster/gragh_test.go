package cluster

import (
	"testing"

	"gopkg.in/check.v1"

	"walm/router/api/v1/instance"
)

func Test(t *testing.T) { check.TestingT(t) }

type clusterSuite struct{}

var _ = check.Suite(&clusterSuite{})

func (cs *clusterSuite) Test_expandDep(c *check.C) {
	leaf := &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test", Links: map[string]string{}},
			Depend: map[string]*instance.Application{
				"test_chart_1:1.0.0": &instance.Application{Name: "test_1"},
				"test_chart_2":       &instance.Application{Name: "test_2"},
			},
			ClusterId: 1001,
		},
	}
	app := expandDep(leaf)
	c.Assert(app.Links["test_chart_1"], check.Equals, "test_1")
	c.Assert(app.Links["test_chart_2"], check.Equals, "test_2")
}

func (cs *clusterSuite) Test_isSameLeaf(c *check.C) {
	a := &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.0"},
		},
	}
	b := &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.0"},
		},
		color: true,
	}

	re := isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

	b = &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.0"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, true)

	b = &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test_2", Chart: "test_chart_1:1.0.0"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

	b = &Leaf{
		app: &Application{
			inst: &instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.1"},
		},
		color: false,
	}

	re = isSameLeaf(a, b)
	c.Assert(re, check.Equals, false)

}

type mockDepInst struct{}

func (mdi *mockDepInst) getDeps(leaf *Leaf) []Leaf {
	if leaf.app.inst.Chart == "test_chart_1:1.0.0" {
		return []Leaf{
			Leaf{
				app: &Application{
					inst:     &instance.Application{Name: "test_1_1", Chart: "test_chart_1_1:1.0.0", Links: map[string]string{}},
					Depend:   map[string]*instance.Application{},
					Bedepend: []*Application{},
				},
			},
			Leaf{
				app: &Application{
					inst:     &instance.Application{Name: "test_1_2", Chart: "test_chart_1_2:1.0.0", Links: map[string]string{}},
					Depend:   map[string]*instance.Application{},
					Bedepend: []*Application{},
				},
			},
		}
	}

	if leaf.app.inst.Chart == "test_chart_2:1.0.0" {
		return []Leaf{
			Leaf{
				app: &Application{
					inst:     &instance.Application{Name: "test_2_1", Chart: "test_chart_2_1:1.0.0", Links: map[string]string{}},
					Depend:   map[string]*instance.Application{},
					Bedepend: []*Application{},
				},
			},
			Leaf{
				app: &Application{
					inst:     &instance.Application{Name: "test_1_2", Chart: "test_chart_1_2:1.0.0", Links: map[string]string{}},
					Depend:   map[string]*instance.Application{},
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
			inst:      &instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.0", Links: map[string]string{}},
			Depend:    map[string]*instance.Application{},
			Bedepend:  []*Application{},
			ClusterId: 1001,
		},
		color: false,
	}

	*a_leaf = append(*a_leaf, *leaf)

	depArray := getDepArray(leaf, a_leaf)

	c.Assert(depArray, check.HasLen, 3)
	c.Assert(depArray[0].app.inst.Name, check.Equals, "test_1")
	c.Assert(depArray[1].app.inst.Name, check.Equals, "test_chart_1_1:1.0.0_1001")
	c.Assert(depArray[2].app.inst.Name, check.Equals, "test_chart_1_2:1.0.0_1001")
	c.Assert(leaf.app.Depend["test_chart_1_1:1.0.0"].Name, check.Equals, "test_chart_1_1:1.0.0_1001")
	c.Assert(leaf.app.Depend["test_chart_1_2:1.0.0"].Name, check.Equals, "test_chart_1_2:1.0.0_1001")
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
		Apps: []instance.Application{
			instance.Application{Name: "test_1", Chart: "test_chart_1:1.0.0", Links: map[string]string{}},
			instance.Application{Name: "test_2", Chart: "test_chart_2:1.0.0", Links: map[string]string{}},
		},
	}
	err, instList := getGragh(1001, "test", "test", cluster)

	c.Assert(err, check.IsNil)
	c.Assert(instList, check.HasLen, 5)
	c.Assert(instList[0].Name, check.Equals, "test_chart_2_1:1.0.0_1001")
	c.Assert(instList[1].Name, check.Equals, "test_chart_1_2:1.0.0_1001")
	c.Assert(instList[2].Name, check.Equals, "test_chart_1_1:1.0.0_1001")
	c.Assert(instList[3].Name, check.Equals, "test_1")
	c.Assert(instList[4].Name, check.Equals, "test_2")

}

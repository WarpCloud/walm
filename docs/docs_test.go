package docs

import (
	"testing"

	"github.com/swaggo/swag"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type docSuite struct {
}

var _ = Suite(&docSuite{})

func (ds *docSuite) TestReadDoc(c *C) {
	doc, _ := swag.ReadDoc()
	c.Assert(doc, Matches, ".*lifecycle.*")
}

package docs

import (
	"testing"

	"github.com/swaggo/swag"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type docSuite struct {
}

var _ = check.Suite(&docSuite{})

func (ds *docSuite) TestReadDoc(c *check.C) {
	doc, _ := swag.ReadDoc()
	c.Assert(doc, check.Matches, "(?s).*cluster lifecycle manager.*")
}

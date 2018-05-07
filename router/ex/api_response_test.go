package ex

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type apiResSuite struct {
}

var _ = Suite(&apiResSuite{})

func (ars *apiResSuite) TestApiResponse_returnBadRequest(c *C) {
	br := ReturnBadRequest()
	c.Assert(br.Code, Equals, http.StatusBadRequest)
	c.Assert(br.Message, Equals, "Invalid Name supplied")
}

func (ars *apiResSuite) TestApiResponse_returnInternalServerError(c *C) {
	err := myerr{msg: "server error!"}
	ise := ReturnInternalServerError(err)

	c.Assert(ise.Code, Equals, http.StatusInternalServerError)
	c.Assert(ise.Message, Equals, "server error!")
}

type myerr struct {
	msg string
}

func (me myerr) Error() string {
	return me.msg
}

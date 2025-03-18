package mapping

import (
	"net/http"

	"go-chassis/src/usecases/api/mapping/parse"

	"github.com/go-chi/render"
)

type Default[in any, out any] struct{}

func (Default[in, out]) ParseRequest(r *http.Request) (*in, error) {
	panic("implement if generic implementation REALLY needed")
}

func (Default[in, out]) HandleReply(res *out, w http.ResponseWriter, r *http.Request) {
	panic("implement if generic implementation REALLY needed")
}

func (Default[in, out]) HandleErr(err error, w http.ResponseWriter, r *http.Request) {
	panic("implement if generic implementation REALLY needed")
}

type AppendToutBox struct {
	Default[any, struct{}]
}

func (AppendToutBox) ParseRequest(r *http.Request) (*any, error) {
	return parse.GenericParse[any](r, true)
}

func (AppendToutBox) HandleReply(res *struct{}, w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

func (AppendToutBox) HandleErr(err error, w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, err.Error())
}

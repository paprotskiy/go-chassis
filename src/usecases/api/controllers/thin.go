package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
)

type ApiParser[In any, Out any] interface {
	ParseRequest(r *http.Request) (*In, error)
	HandleReply(res *Out, w http.ResponseWriter, r *http.Request)
	HandleErr(err error, w http.ResponseWriter, r *http.Request)
}

func Thin[In any, Out any](api ApiParser[In, Out], operation func(context.Context, In) (*Out, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := api.ParseRequest(r)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err.Error())
			return
		}

		res, err := operation(r.Context(), *input)
		if err != nil {
			api.HandleErr(err, w, r)
			return
		}

		api.HandleReply(res, w, r)
	}
}

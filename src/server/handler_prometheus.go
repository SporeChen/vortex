package server

import (
	"fmt"
	"net/http"
	"time"

	restful "github.com/emicklei/go-restful"
	response "github.com/linkernetworks/vortex/src/net/http"
	"github.com/linkernetworks/vortex/src/net/http/query"
	"github.com/linkernetworks/vortex/src/serviceprovider"
	"github.com/linkernetworks/vortex/src/web"
	"github.com/prometheus/common/model"
	"golang.org/x/net/context"
)

func queryMetrics(ctx *web.Context) {
	sp, req, resp := ctx.ServiceProvider, ctx.Request, ctx.Response

	query := query.New(req.Request.URL.Query())

	var expression string
	if q, ok := query.Str("query"); ok {
		expression = q
	} else {
		response.BadRequest(req.Request, resp.ResponseWriter, fmt.Errorf("wrong expression: %v", q))
	}

	result, err := queryFromPrometheus(sp, expression)
	if err != nil {
		response.BadRequest(req.Request, resp.ResponseWriter, err)
	}

	resp.WriteJson(map[string]interface{}{
		"status":  http.StatusOK,
		"results": result,
	}, restful.MIME_JSON)
}

func queryFromPrometheus(sp *serviceprovider.Container, expression string) (model.Vector, error) {

	api := sp.Prometheus.API

	testTime := time.Now()
	result, err := api.Query(context.Background(), expression, testTime)

	// https://github.com/prometheus/client_golang/blob/d6a9817c4afc94d51115e4a30d449056a3fbf547/api/prometheus/v1/api.go#L316
	// this api always return the err no matter what
	// so we should use result==nil to determine whether it is a true error
	if result == nil {
		return nil, err
	}

	switch {
	case result.Type() == model.ValVector:
		return result.(model.Vector), nil
	default:
		return nil, fmt.Errorf("the type of the return result can not be identify")
	}
}

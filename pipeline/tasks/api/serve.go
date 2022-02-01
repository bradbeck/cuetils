package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
  "sync"

	"cuelang.org/go/cue"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
  "github.com/labstack/echo-contrib/prometheus"

  "github.com/hofstadter-io/cuetils/pipeline/context"
  "github.com/hofstadter-io/cuetils/pipeline/pipe"
)

func init() {
  context.Register("api.Serve", NewServe)
}

type Serve struct {
  sync.Mutex
}

func NewServe(val cue.Value) (context.Runner, error) {
  return &Serve{}, nil
}

func (T *Serve) Run(ctx *context.Context) (interface{}, error) {
  var err error

  val := ctx.Value

  logging := false
  l := val.LookupPath(cue.ParsePath("logging"))
  if l.Exists() {
    if l.Err() != nil {
      return nil, l.Err()
    }
    logging, err = l.Bool()
    if err != nil {
      return nil, err
    }
  }

  // get the port
  p := val.LookupPath(cue.ParsePath("port"))
  if p.Err() != nil {
    return nil, p.Err()
  }
  port, err := p.String()
  if err != nil {
    return nil, err
  }

  // create server
  e := echo.New()
  e.HideBanner = true
  e.Use(middleware.Recover())
  if logging {
    e.Use(middleware.Logger())
  }

  // liveliness and metrics
	e.GET("/alive", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

  prom := prometheus.NewPrometheus("echo", nil)
  prom.Use(e)

  //
  // Setup routes
  //
  routes := val.LookupPath(cue.ParsePath("routes"))
  iter, err := routes.Fields()
  if err != nil {
    return nil, err
  }

  for iter.Next() {
    label := iter.Selector().String()
    route := iter.Value()

    // fmt.Println("route:", label)

    err := T.routeFromValue(label, route, e, ctx)
    if err != nil {
      return nil, err
    }
  }

  /*
  // print routes
  data, err := json.MarshalIndent(e.Routes(), "", "  ")
  if err != nil {
    return err
  }

  fmt.Println(string(data))
  */

  // run the server
	e.Logger.Fatal(e.Start(":" + port))

  fmt.Println("SERVER EXITED")

  // - pull apart server value

  /*
  - loop over routes...
    - should be a pipeline, we need to load this
    - construct the routes, with pipeline & echo framework
  - go run the server
  - wait for exit signal


  */
	return nil, err
}

func (T *Serve) routeFromValue(path string, route cue.Value, e *echo.Echo, ctx *context.Context) (error) {
  path = strings.Replace(path, "\"", "", -1)
  // fmt.Println(path + ":", route)

  // is this a pipeline handler?
  attrs := route.Attributes(cue.ValueAttr)
  isPipe := false
  for _, a := range attrs {
    if a.Name() == "pipeline" {
      isPipe = true
    } 
  }

  local := route

  // fmt.Println("setting up route:", path, isPipe)

  // (1) can we read the pipelinen once and reuse it
  // (2) or do we need to construct a new one on each call

  // setup handler, this will be invoked on all requests
  handler := func (c echo.Context) error {
    // fmt.Println("handling:", path)
    // pull apart c.request
    req, err := T.buildReqValue(c)
    if err != nil {
      return err
    }
    // fmt.Println("reqVal", req)

    tmp := local.FillPath(cue.ParsePath("req"), req)
    if tmp.Err() != nil {
      return tmp.Err()
    }

    if isPipe {
      p, err := pipe.NewPipeline(ctx, tmp)
      if err != nil {
        return err
      }

      err = p.Start()
      if err != nil {
        return err
      }

      tmp = p.Final
    }

    resp := tmp.LookupPath(cue.ParsePath("resp"))
    if resp.Err() != nil {
      return resp.Err()
    }

    err = T.fillRespFromValue(resp, c)
    if err != nil {
      return err
    }

    return nil
  }

  // figure out route method(s): GET, POST, et al
  mv := route.LookupPath(cue.ParsePath("method"))
  methods := []string{}
  switch mv.IncompleteKind() {
  case cue.StringKind:
    m, err := mv.String()
    if err != nil {
      return err
    }
    m = strings.ToUpper(m)
    methods = append(methods, m)
  case cue.ListKind:
    iter, err := mv.List()
    if err != nil {
      return err
    }
    for iter.Next() {
      v := iter.Value()
      m, err := v.String()
      if err != nil {
        return err
      }
      m = strings.ToUpper(m)
      methods = append(methods, m)
    }

  case cue.BottomKind:
    methods = append(methods, "GET")

  default: 
    return fmt.Errorf("unsupported type for method in %s %v", path, mv.IncompleteKind)
  }

  // fmt.Println("methods:", methods)
  e.Match(methods, path, handler)

  return nil
}

func (T *Serve) buildReqValue(c echo.Context) (interface{},error) {
  req := map[string]interface{}{}
  R := c.Request()

  req["method"] = R.Method
  req["header"] = R.Header
  req["url"] = R.URL
  req["query"] = c.QueryParams()

  b, err := io.ReadAll(R.Body)
  if err != nil {
    return nil, err
  }

  if len(b) > 0 {
    var body interface{}
    err = json.Unmarshal(b, &body)
    if err != nil {
      return nil, err
    }
    req["body"] = body
  }

  // form
  // path params
  return req, nil
}

func (T *Serve) fillRespFromValue(val cue.Value, c echo.Context) error {
  var ret map[string]interface{}

  {
    T.Lock()
    defer T.Unlock()

    err := val.Decode(&ret)
    if err != nil {
      return err
    }
  }

  // TODO, more http/response type things

  st, ok := ret["status"]
  if !ok {
    st = 200
  }
  status := st.(int)

  if ret["json"] != nil {
    return c.JSON(status, ret["json"])
  } else if ret["html"] != nil {
    // todo, better type casts
    return c.HTML(status, ret["html"].(string))
  } else if ret["body"] != nil {
    // todo, better type casts
    return c.String(status, ret["body"].(string))
  } else {
    return c.NoContent(status)
  }
}

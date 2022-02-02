package os

import (
  "bufio"
  "fmt"

	"cuelang.org/go/cue"

  "github.com/hofstadter-io/cuetils/flow/context"
)

func init() {
  context.Register("os.Stdin", NewStdin)
}

type Stdin struct {}

func NewStdin(val cue.Value) (context.Runner, error) {
  return &Stdin{}, nil
}

func (T *Stdin) Run(ctx *context.Context) (interface{}, error) {

	v := ctx.Value
  var m string

  ferr := func () error {
    ctx.CUELock.Lock()
    defer func() {
      ctx.CUELock.Unlock()
    }()
    var err error

    msg := v.LookupPath(cue.ParsePath("msg")) 
    if msg.Err() != nil {
      return msg.Err()

    } else if msg.Exists() {
      m, err = msg.String()
      if err != nil {
        return err
      }
      // print message to user
    }
    return nil
  }()
  if ferr != nil {
    return nil, ferr
  }

  if len(m) > 0 {
    fmt.Fprint(ctx.Stdout, m)
  }
  reader := bufio.NewReader(ctx.Stdin)
  text, _ := reader.ReadString('\n')

  ctx.CUELock.Lock()
  defer ctx.CUELock.Unlock()
  res := v.FillPath(cue.ParsePath("contents"), text)

	return res, nil 
}

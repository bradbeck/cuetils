package msg

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"cuelang.org/go/cue"
	"gopkg.in/irc.v3"

  "github.com/hofstadter-io/cuetils/pipeline/context"
  "github.com/hofstadter-io/cuetils/pipeline/pipe"
	"github.com/hofstadter-io/cuetils/utils"
)

func init() {
  context.Register("msg.IrcClient", NewIrcClient)
}

type IrcClient struct{}

func NewIrcClient(val cue.Value) (context.Runner, error) {
  return &IrcClient{}, nil
}

func (T *IrcClient) Run(ctx *context.Context) (interface{}, error) {

	val := ctx.Value

  config, err := buildIrcConfig(val)
	if err != nil {
    fmt.Println("irc: buildConfig err:", err)
    return nil, err	
	}

  handler, err := buildIrcHandler(ctx, val)
	if err != nil {
    fmt.Println("irc: buildHandler err:", err)
    return nil, err	
	}

  config.Handler = handler

  h := val.LookupPath(cue.ParsePath("host"))
  if h.Err() != nil {
    return nil, h.Err()
  }
  host, err := h.String()
  if err != nil {
    return nil, err
  }
	conn, err := net.Dial("tcp", host)
	if err != nil {
    return nil, err	
	}

	// Create the client
	client := irc.NewClient(conn, config)

  var wg sync.WaitGroup

  wg.Add(1)
  go func() {
    defer wg.Done()

    fmt.Println("For realz this time")
    err = client.Run()
    if err != nil {
      fmt.Println("IRC err:", err)
      return	
    }
    fmt.Println("irc done")
  }()

  wg.Wait()
  fmt.Println("ending", err)
  return nil, err
}

func buildIrcConfig(val cue.Value) (irc.ClientConfig, error) {
	config := irc.ClientConfig{}

  n := val.LookupPath(cue.ParsePath("nick"))
  if n.Err() != nil {
    return config, n.Err()
  }
  nick, err := n.String()
  if err != nil {
    return config, err
  }
  config.Nick = nick

  p := val.LookupPath(cue.ParsePath("pass"))
  if p.Err() != nil {
    return config, p.Err()
  }
  pass, err := p.String()
  if err != nil {
    return config, err
  }
  config.Pass = pass

  return config, nil
}

func buildIrcHandler(ct_ctx *context.Context, val cue.Value) (irc.HandlerFunc, error) {
  fmt.Println("Building IRC handler:")
  ctx := val.Context()

  c := val.LookupPath(cue.ParsePath("channel"))
  if c.Err() != nil {
    return nil, c.Err()
  }
  channel, err := c.String()
  if err != nil {
    return nil, err
  }

  lM := val.LookupPath(cue.ParsePath("log_msgs"))
  if lM.Err() != nil {
    return nil, lM.Err()
  }
  logMsgs, err := lM.Bool()
  if err != nil {
    return nil, err
  }

  cHandler := val.LookupPath(cue.ParsePath("handler"))
  if !cHandler.Exists() {
    fmt.Println("got here")
    return nil, cHandler.Err()
  }

  fmt.Println("handler:", cHandler)

  handler := func(c *irc.Client, m *irc.Message) {
    // turn incoming msg into a cue.Value
    bs, err := json.Marshal(m)
    if err != nil {
      fmt.Println("Error(json):", err)
      return
    }

    mv := ctx.CompileBytes(bs)
    if mv.Err() != nil {
      fmt.Println("Error(cuepile):", mv.Err())
      return
    }

    ms, err := utils.PrintCue(mv)
    if err != nil {
      fmt.Println("Error(print):", err)
      return
    }

    if logMsgs {
      fmt.Println(ms)
    }

    // some shortcut reposonse for all IRC
    switch m.Command {
    case "PING":
      // do we need a pong config value?
      // hopefully not, as one would assume this has been standardized
      host := m.Params[0]
      c.Write("PONG " + host)
      return
    case "001":
      msgs := val.LookupPath(cue.ParsePath("init_msgs"))
      iter, err := msgs.List()
      if err != nil {
        fmt.Printf("Error: IRC.init_msgs should be a list of strings\n%v\n", err)
      }

      for iter.Next() {
        sv := iter.Value()
        s, err := sv.String()
        if err != nil {
          fmt.Printf("Error: IRC.init_msgs should be a list of strings\n%v\n", err)
        }
        fmt.Println("sending(init):", s)
        c.Write(s)
      } 

      return
    }




    v := ctx.CompileString("{...}")
    v = v.Unify(cHandler) 
    v = v.FillPath(cue.ParsePath("msg"), mv)

    // is this a pipeline
    errV := v.LookupPath(cue.ParsePath("error"))
    respV := v.LookupPath(cue.ParsePath("resp"))
    pipeV := v.LookupPath(cue.ParsePath("pipe"))

    fmt.Println("errV:", errV)
    fmt.Println("respV:", respV)
    fmt.Println("pipeV:", pipeV)

    // log any errors
    if errV.Exists() {
      fmt.Println("Error in handler:", errV)
      return
    }

    // write back simple responses
    if respV.Exists() {
      fmt.Println("found respV")
      s, err := respV.String()
      if err != nil {
        fmt.Println("Error(respV):", err)
      }

      fmt.Println("sending(msg):", s)
      c.Writef("PRIVMSG %s :%s", channel, s)
      return
    }

    // handle pipelines
    if pipeV.Exists() {
      // build new value
      v := ctx.CompileString("{...}")
      v = v.Unify(pipeV) 

      p, err := pipe.NewPipeline(ct_ctx, v)
      if err != nil {
        fmt.Println("Error(pipe/new):", err)
        return
      }

      err = p.Start()
      if err != nil {
        fmt.Println("Error(pipe/run):", err)
        return
      }

      rV := p.Final.LookupPath(cue.ParsePath("resp"))
      if !rV.Exists() {
        fmt.Println("Error(pipe/resp): does not exist")
        return 
      }
      s, err := rV.String()
      if err != nil {
        fmt.Println("Error(pipe/rVstr):", err)
        return
      }

      // fill in go-irc.Message and then turn that into a string

      fmt.Println("sending(pipe/msg):", s)
      c.Writef("PRIVMSG %s :%s", channel, s)

      return
      
    }

    // otherwise, unknown message
    fmt.Println("unhandled message:", ms)


  }

  return handler, nil
}

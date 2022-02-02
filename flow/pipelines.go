package flow

import (
  "fmt"

  "cuelang.org/go/cue"

	"github.com/hofstadter-io/cuetils/cmd/cuetils/flags"
	"github.com/hofstadter-io/cuetils/flow/context"
	"github.com/hofstadter-io/cuetils/flow/pipe"
	"github.com/hofstadter-io/cuetils/structural"
	"github.com/hofstadter-io/cuetils/utils"
)

func hasFlowAttr(val cue.Value, args []string) (attr cue.Attribute, found, keep bool) {
  attrs := val.Attributes(cue.ValueAttr)

  for _, attr := range attrs {
    if attr.Name() == "flow" {
      // found a flow, stop recursion
      found = true
      // if it matches our args, create and append
      keep = matchFlow(attr, args)
      if keep {
        return attr, true, true
      }
    }
  }

  return cue.Attribute{}, found, false
}

func matchFlow(attr cue.Attribute, args []string) (keep bool) {
  // fmt.Println("matching 1:", attr, args, len(args), attr.NumArgs())
  // if no args, match flows without args
  if len(args) == 0 {
    if attr.NumArgs() == 0 {
      return true
    }
    // extra check for one arg which is empty
    if attr.NumArgs() == 1 {
      s, err := attr.String(0)
      if err != nil {
        fmt.Println("bad flow tag:", err)
        return false
      }
      return s == ""
    }

    return false
  }

  // for now, match any
  // upgrade logic for user later
  for _, tag := range args {
    for p := 0; p < attr.NumArgs(); p++ {
      s, err := attr.String(p)
      if err != nil {
        fmt.Println("bad flow tag:", err)
        return false
      }
      if s == tag {
        return true
      }
    }
  }

  return false
}

func listFlows(val cue.Value,  opts *flags.RootPflagpole, popts *flags.FlowFlagpole) (error) {
  args := popts.Flow

  printer := func(v cue.Value) bool {
    attrs := v.Attributes(cue.ValueAttr)

    for _, attr := range attrs {
      if attr.Name() == "flow" {
        if len(args) == 0 || matchFlow(attr, args) {
          if popts.Docs {
            s := ""
            docs := v.Doc()
            for _, d := range docs {
              s += d.Text()
            }
            fmt.Print(s)
          }
          if opts.Verbose {
            s, _ := utils.FormatCue(v)
            fmt.Printf("%s: %s\n", v.Path(), s)
          } else {
            fmt.Println(attr)
          }
        }
        return false
      }
    }

    return true
  }

  structural.Walk(val, printer, nil, walkOptions...)

  return nil
}

// maybe this becomes recursive so we can find
// nested flows, but not recurse when we find one
// only recurse when we have something which is not a flow or task
func findFlows(ctx *context.Context, val cue.Value, opts *flags.RootPflagpole, popts *flags.FlowFlagpole) ([]*pipe.Flow, error) {
  pipes := []*pipe.Flow{}

  // TODO, look for lists?
  s, err := val.Struct()
  if err != nil {
    // not a struct, so don't recurse
    // fmt.Println("not a struct", err)
    return nil, nil
  }

  args := popts.Flow

  // does our top-level (file-level) have @flow()
  _, found, keep := hasFlowAttr(val, args)
  if keep  {
    // invoke TaskFactory
    p, err := pipe.NewFlow(ctx, val)
    if err != nil {
      return pipes, err
    }
    pipes = append(pipes, p)
  }

  if found {
    return pipes, nil
  }

  iter := s.Fields(
		cue.Attributes(true),
		cue.Docs(true),
  )

  // loop over top-level fields in the cue instance
  for iter.Next() {
    v := iter.Value()

    _, found, keep := hasFlowAttr(v, args)
    if keep  {
      p, err := pipe.NewFlow(ctx, v)
      if err != nil {
        return pipes, err
      }
      pipes = append(pipes, p)
    }

    // break recursion if flow found
    if found {
      continue
    }

    // recurse to search for more flows
    ps, err := findFlows(ctx, v, opts, popts)
    if err != nil {
      return pipes, nil 
    }
    pipes = append(pipes, ps...)
  }

  return pipes, nil
}


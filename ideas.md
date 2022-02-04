# Today

- [ ] implement oauth workflow for youtube
- [ ] update twitch library from learnings with youtube

---

once tired...
- see if we can `cue fmt` on buffer save

---

- [x] try CUE v0.4.2-rc1, need to push hof patch
- [ ] youtube in streamer-tools

- flows
  - [ ] new 'flows' repo in hof github as a
        place to put reusable and example tasks&flows)
  - [ ] docker containers
    - [ ] task library for importing
    - [ ] start / stop

---

at some point...
migrate to hof/flow cmd

### Tests...

txtar in 

### Docs...

- [ ] start hof/docs folder(s)
- [ ] task schemas (github.com/hof-io/hof/flow/tasks/...)
- [ ] task reference (autogen the majority)

### hof/flow examples

- [ ] schemas for builtin tasks
- [ ] make this list
  - [ ] examples for available tasks
  - [ ] composite tasks like docker
  - [ ] links to tools using flow (streamer-tools)

- save all IRC messages to DB
- bookmarks and hn upvotes
- change my lights
- replace helm (need native topo sort)
- OAuth workflow

### hof/flow internals

- [ ] metrics and progress
  - [ ] TaskStats: struct and centrally collecting
  - [ ] Print progress (and stats) by flag
  - [ ] failure mode for tasks, some common schema across all tasks? @onfail()

- [ ] i/o centralization
  - [ ] debug/verbose flag to print tasks which run
  - [ ] stats for tasks & pipelines, chan to central
  - [ ] obfescate secrets, centralized printing (ensure that is the case in every task / run)

- [ ] exec improvements
  - [ ] many options not enabled yet
  - [ ] for exec: pipe stdin/out/err to files
  - [ ] some way to run in background, and then kill / exit later?

- [ ] async / client listener
  - [ ] chan / mailbox
  - [ ] waitgroup / mutex?
  - [ ] kill chan, also need to catch signals in main(?) and pass down / do right thing
    - [ ] how to tell (server / bg exec'd process) to stop (oauth localhost as example)
  - [ ] websockets

- something that loops over input list and produces messages / tasks (?)

- [ ] sql
  - [x] sqlite
  - [ ] postgres
  - [ ] mysql

- [ ] msg
  - [ ] rabbitmq
  - [ ] kafka
  - [ ] nats

- [ ] k/v
  - [ ] redis
  - [ ] memcache
  - [ ] gcs
  - [ ] s3

- [ ] obj
  - [ ] elasticsearch
  - [ ] mongo

- [ ] mouse/keyboard automation
  - [ ] Browswer - https://github.com/playwright-community/playwright-go
  - [ ] OS level - https://github.com/go-vgo/robotgo


Other: 

- specify CLI args to flow for command to run (avoid -p)
- better (boolean,regexp) logic for selecting pipeline(s)
- temp files / dirs
- command line prompt
- support for fs.FS (https://github.com/hack-pad/hackpadfs)
  - could abstract away S3/GS

### Build other things cuetils/run

### More todo, always...

Exec & Serve & async

- [ ] write directly to file for stdio, is it a concrete string?
- [ ] something like a goroutine, similar to api.Serve/pipeline
- [ ] message passing, via chans, websockets, kafka/rabbit

Bugs?

- [ ] prevent exit when error in handler pipelines?
- [ ] rare & racey structural cycle
- [ ] cuetils flow args for CWD all behave differently
  - [ ] `<no args>` -> no input
  - [ ] `*.cue` -> refs across files not found
  - [ ] `./` -> works
- [ ] imported flows that have typos don't throw error, rather they don't show up silently

Helpers:

- canonicalize (sort fields recursively)
- toposort

List processing:

- jsonl
- yaml with `---`
- CUE got streaming json/yaml support
- if extracted value is a list?

### Other

Go funcs:

- rename currenct to `*Globs`
- pure Go implementations
- funcs that take values

CLI:

- Support expression on globs, to select out a field on each file
- move implementation?

### futurology

- @label(), but also part of evaluation? (available for gens and flow)
- diff lists, @id(), how to detect renames and position changes and optimize?


## upstream & references

#### Memory issues

(we have not seen this yet with the twitch IRC bot which had lots of activity)

https://github.com/lipence/cue/commit/6ed69100ebd62509577826657536172ab46cf257

#### cue/flow

return final value: https://github.com/cue-lang/cue/pull/1390

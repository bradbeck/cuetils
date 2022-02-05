package hof

import "strings"

compile: {
  @flow(compile)

  reporoot: { 
    @task(os.Exec)
    cmd: ["bash", "-c", "git rev-parse --show-toplevel"]
    stdout: string
    out: strings.TrimSpace(stdout)
  }

  root: reporoot.out
  dirs: ["cmd","flow","recurse","structural","utils"]

  watch: {
    @task(fs.Watch)
    globs: [ for d in dirs { "\(root)/\(d)/**/*.go" } ]
    handler: {
      event?: _
      compile: {
        @task(os.Exec)
        cmd: ["go", "install", "\(root)/cmd/cuetils"]
        stdout: string
      }
    }
  }
}

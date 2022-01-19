req: {
	host: "https://postman-echo.com"
	method: "GET"
	path: "/get"
	query: {
		cow: "moo"
	}
}

pick: {
	args: cow: string
}

tasks: [string]: {
	Out: _
	...
}

tasks: {
  @pipeline(api-test)
	r1: { #Req: req, Resp: _ } @task(api/call) @print("#Req",Resp)
	p1: { #X: r1.Resp, #P: pick } @task(st/pick) @print(Out)
}

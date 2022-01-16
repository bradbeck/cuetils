x: {
	"a": {
		"b": "B"
	},
	"b": 1,
	"c": 2,
	"d": "D"
}

y: {
	a: {
		b: string
	}
	c: int
	d: "D"
}

tasks: [string]: {
	Out: _
	...
}

tasks: {

	p1: { #X: x, #P: y } @pick()
	
	m1: { #X: p1.Out, #M: { c: int } } @mask()
	m2: { #X: p1.Out, #M: { a: _ } } @mask()

	u1: { #X: m1.Out, #U: m2.Out } @upsert()
}

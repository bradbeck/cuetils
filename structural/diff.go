package structural

import (
	"cuelang.org/go/cue"

	"github.com/hofstadter-io/cuetils/cmd/cuetils/flags"
)

func DiffGlobs(orig string, next string, opts *flags.RootPflagpole) ([]GlobResult, error) {
	return BinaryOpGlobs(orig, []string{next}, opts, DiffValue)
}

func DiffValue(orig, next cue.Value, opts *flags.RootPflagpole) (cue.Value, error) {
	r, _ := diffValue(orig, next, opts)
	return r, nil
}

func diffValue(orig, next cue.Value, opts *flags.RootPflagpole) (cue.Value, bool) {
	switch orig.IncompleteKind() {
	case cue.StructKind:
		// fmt.Println("struct", orig, next)
		return diffStruct(orig, next, opts)

	case cue.ListKind:
		// fmt.Println("list", orig, next)
		return diffList(orig, next, opts)

	default:
		// fmt.Println("leaf", orig, next)
		return diffLeaf(orig, next, opts)
	}
}

func diffStruct(orig, next cue.Value, opts *flags.RootPflagpole) (cue.Value, bool) {
	ctx := orig.Context()
	result := newStruct(ctx)
	add := newStruct(ctx)
	rmv := newStruct(ctx)
	didAdd := false
	didRmv := false

	// first loop over val
	iter, _ := orig.Fields(defaultWalkOptions...)
	for iter.Next() {
		s := iter.Selector()
		p := cue.MakePath(s)
		u := next.LookupPath(p)

		// check that field exists in from. Should we be checking f.Err()?
		if u.Exists() {
			r, ok := diffValue(iter.Value(), u, opts)
			// fmt.Println("r:", r, ok, p)
			if ok {
				result = result.FillPath(p, r)
			}
		} else {
			// remove if orig not in next
			didRmv = true
			rmv = rmv.FillPath(p, iter.Value())
		}
	}

	// add anything in next that is not in orig
	iter, _ = next.Fields(defaultWalkOptions...)
	for iter.Next() {
		s := iter.Selector()
		p := cue.MakePath(s)
		v := orig.LookupPath(p)

		// check that field exists in from. Should we be checking f.Err()?
		if !v.Exists() {
			didAdd = true
			add = add.FillPath(p, iter.Value())
		}
	}

	if didRmv {
		result = result.FillPath(cue.ParsePath("\"-\""), rmv)
	}
	if didAdd {
		result = result.FillPath(cue.ParsePath("\"+\""), add)
	}

	return result, true
}

func diffList(orig, next cue.Value, opts *flags.RootPflagpole) (cue.Value, bool) {
	ctx := orig.Context()
	oi, _ := orig.List()
	ni, _ := next.List()

	result := []cue.Value{}
	for oi.Next() && ni.Next() {
		v, ok := pickValue(oi.Value(), ni.Value(), opts)
		if ok {
			result = append(result, v)
		}
	}

	return ctx.NewList(result...), true
}

func diffLeaf(orig, next cue.Value, opts *flags.RootPflagpole) (cue.Value, bool) {
	ctx := orig.Context()
	ret := newStruct(ctx)
	lbl := GetLabel(orig)

	// check if they are the same type and concreteness, and check if unify, if so, no need to add to diff
	if orig.IncompleteKind() == next.IncompleteKind() {
		if orig.IsConcrete() == next.IsConcrete() {
			u := orig.Unify(next)
			if u.Err() == nil {
				return ret, false
			}
		}
	}

	// otherwise, we have a diff to create
	rmv := newStruct(ctx)
	rmv = rmv.FillPath(cue.MakePath(lbl), orig)
	ret = ret.FillPath(cue.ParsePath("\"-\""), rmv)

	add := newStruct(ctx)
	add = add.FillPath(cue.MakePath(lbl), next)
	ret = ret.FillPath(cue.ParsePath("\"+\""), add)

	return ret, true
}

// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package stack

import (
	"reflect"
	"sort"
)

// Similarity is the level at which two call lines arguments must match to be
// considered similar enough to coalesce them.
type Similarity int


type callstack []string
type Callstacks []*callstack

const (
	// ExactFlags requires same bits (e.g. Locked).
	ExactFlags Similarity = iota
	// ExactLines requests the exact same arguments on the call line.
	ExactLines
	// AnyPointer considers different pointers a similar call line.
	AnyPointer
	// AnyValue accepts any value as similar call line.
	AnyValue
)

// Aggregate merges similar goroutines into buckets.
//
// The buckets are ordered in library provided order of relevancy. You can
// reorder at your chosing.
func Aggregate(goroutines []*Goroutine, similar Similarity) []*Bucket {
	type count struct {
		ids   []int
		first bool
	}
	b := map[*Signature]*count{}
	// O(n²). Fix eventually.
	for _, routine := range goroutines {
		found := false
		for key, c := range b {
			// When a match is found, this effectively drops the other goroutine ID.
			if key.similar(&routine.Signature, similar) {
				found = true
				c.ids = append(c.ids, routine.ID)
				c.first = c.first || routine.First
				if !key.equal(&routine.Signature) {
					// Almost but not quite equal. There's different pointers passed
					// around but the same values. Zap out the different values.
					newKey := key.merge(&routine.Signature)
					b[newKey] = c
					delete(b, key)
				}
				break
			}
		}
		if !found {
			// Create a copy of the Signature, since it will be mutated.
			key := &Signature{}
			*key = routine.Signature
			b[key] = &count{ids: []int{routine.ID}, first: routine.First}
		}
	}
	out := make(buckets, 0, len(b))
	for signature, c := range b {
		sort.Ints(c.ids)
		out = append(out, &Bucket{Signature: *signature, IDs: c.ids, First: c.first})
	}
	sort.Sort(out)
	return out
}


/* AggreateSubsets aggregates all subsets of goroutines[] into their toplevel stacks.
 First cut compares every stack to ever other stack. Optimize in due time. */
func AggregateSubsets(goroutines []*Goroutine, allStacks Callstacks) Callstacks {
	if allStacks == nil {
		allStacks = make(Callstacks, 0)
	}
	var stacks []*callstack
	for _, routine := range goroutines {
		stacks = append(stacks, flattenStack(routine.Stack.Calls))
	}
	for _, newstack := range stacks {
		// Modify allstacks by adding/removing the necessary stack.
		allStacks = checkSubset(allStacks, *newstack)
	}
	return allStacks
}

func checkSubset(fullStacks []*callstack, curstack callstack) []*callstack {
	var subset bool
	removeIndexes := make(map[int]bool)
	for i, st := range fullStacks {
		// First check for duplicate
		if reflect.DeepEqual(*st, curstack) {
			subset = true
			break
		} else if isOrderedSubset(&curstack, st) {
			subset = true
			break
		} else if isOrderedSubset(st, &curstack) {
			// The current stack is bigger, keep that instead, remove this one.
			removeIndexes[i] = true
		}
	}
	fullStacksCopy := make([]*callstack, 0)
	for i, st := range fullStacks {
		if _, present := removeIndexes[i]; !present {
			fullStacksCopy = append(fullStacksCopy, st)
		}
	}
	fullStacks = fullStacksCopy
	if !subset {
		fullStacks = append(fullStacks, &curstack)
	}
	return fullStacks
}

// Returns true if first is a subset of second.
func isOrderedSubset(first, second *callstack) bool {
	if len(*first) > len(*second) {
		return false
	}
	set := make(map[string]int)
	for _, value := range *second {
		set[value] += 1
	}

	for _, value := range *first {
		if count, found := set[value]; !found {
			return false
		} else if count < 1 {
			return false
		} else {
			set[value] = count - 1
		}
	}
	return checkSequence(*first, *second)
}

func checkSequence(a, b []string) bool {
	for i:=0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func flattenStack(callStack []Call) *callstack {
	var callList callstack
	for _, call := range callStack {
		callList = append(callList, call.Func.Raw)
	}
	return &callList
}

// Bucket is a stack trace signature and the list of goroutines that fits this
// signature.
type Bucket struct {
	Signature
	// IDs is the ID of each Goroutine with this Signature.
	IDs []int
	// First is true if this Bucket contains the first goroutine, e.g. the one
	// Signature that likely generated the panic() call, if any.
	First bool
}

// less does reverse sort.
func (b *Bucket) less(r *Bucket) bool {
	if b.First || r.First {
		return b.First
	}
	return b.Signature.less(&r.Signature)
}

//

// buckets is a list of Bucket sorted by repeation count.
type buckets []*Bucket

func (b buckets) Len() int {
	return len(b)
}

func (b buckets) Less(i, j int) bool {
	return b[i].less(b[j])
}

func (b buckets) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

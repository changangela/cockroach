// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// {{/*
// +build execgen_template
//
// This file is the execgen template for hashjoiner.eg.go. It's formatted in a
// special way, so it's both valid Go and a valid text/template input. This
// permits editing this file with editor support.
//
// */}}

package exec

import (
	"bytes"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/exec/types"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

// {{/*

// Dummy import to pull in "tree" package.
var _ tree.Datum

// Dummy import to pull in "math" package
var _ = math.Pi

// Dummy import to pull in "bytes" package.
var _ bytes.Buffer

// _ASSIGN_HASH is the template equality function for assigning the first input
// to the result of the hash value of the second input.
func _ASSIGN_HASH(_, _ string) uint64 {
	panic("")
}

func _CHECK_COL_MAIN(_SEL_IND uint64) { // */}}
	// {{define "checkColMain"}}
	buildVal := buildKeys[keyID-1]
	probeVal := probeKeys[_SEL_IND]
	var unique bool
	_ASSIGN_NE("unique", "buildVal", "probeVal")

	if unique {
		prober.differs[prober.toCheck[i]] = true
	}
	// {{end}}

	// {{/*
}

func _CHECK_COL_BODY(_SEL_IND uint64, _PROBE_HAS_NULLS bool, BUILD_HAS_NULLS bool) { // */}}
	// {{define "checkColBody"}}
	for i := uint16(0); i < nToCheck; i++ {
		// keyID of 0 is reserved to represent the end of the next chain.

		if keyID := prober.groupID[prober.toCheck[i]]; keyID != 0 {
			// the build table key (calculated using keys[keyID - 1] = key) is
			// compared to the corresponding probe table to determine if a match is
			// found.

			/* {{if .ProbeHasNulls }} */
			if probeVec.NullAt(_SEL_IND) {
				prober.groupID[prober.toCheck[i]] = 0
			} else /*{{end}} {{if .BuildHasNulls}} */ if buildVec.NullAt64(keyID - 1) {
				prober.differs[prober.toCheck[i]] = true
			} else /*{{end}} */ {
				_CHECK_COL_MAIN(_SEL_IND)
			}
		}
	}
	// {{end}}
	// {{/*
}

func _CHECK_COL_WITH_NULLS(_SEL_IND uint64) { // */}}
	// {{define "checkColWithNulls"}}
	if probeVec.HasNulls() {
		if buildVec.HasNulls() {
			_CHECK_COL_BODY(_SEL_IND, true, true)
		} else {
			_CHECK_COL_BODY(_SEL_IND, true, false)
		}
	} else {
		if buildVec.HasNulls() {
			_CHECK_COL_BODY(_SEL_IND, false, true)
		} else {
			_CHECK_COL_BODY(_SEL_IND, false, false)
		}
	}
	// {{end}}
	// {{/*
}

func _REHASH_BODY(_ string) { // */}}
	// {{define "rehashBody"}}
	for i := uint64(0); i < nKeys; i++ {
		v := keys[_SEL_IND]
		var hash uint64
		_ASSIGN_HASH("hash", "v")
		buckets[i] = buckets[i]*31 + hash
	}
	// {{end}}

	// {{/*
}

func _COLLECT_RIGHT_OUTER(_SEL_IND uint64) { // */}}
	// {{define "collectRightOuter"}}
	for i := uint16(0); i < batchSize; i++ {
		currentID := prober.head[i]

		if currentID == 0 {
			prober.buildNil[nResults] = true
		}

		for {
			if nResults >= ColBatchSize {
				prober.prevBatch = batch
				return nResults
			}

			prober.buildIdx[nResults] = currentID - 1
			prober.probeIdx[nResults] = _SEL_IND
			currentID = prober.ht.same[currentID]
			prober.head[i] = currentID
			nResults++

			if currentID == 0 {
				break
			}
		}
	}
	// {{end}}
	// {{/*
}

func _COLLECT_NO_OUTER(_SEL_IND uint64) { // */}}
	// {{define "collectNoOuter"}}
	for i := uint16(0); i < batchSize; i++ {
		currentID := prober.head[i]
		for currentID != 0 {
			if nResults >= ColBatchSize {
				prober.prevBatch = batch
				return nResults
			}

			prober.buildIdx[nResults] = currentID - 1
			prober.probeIdx[nResults] = _SEL_IND
			currentID = prober.ht.same[currentID]
			prober.head[i] = currentID
			nResults++
		}
	}
	// {{end}}
	// {{/*
}

func _DISTINCT_COLLECT_RIGHT_OUTER(_SEL_IND uint64) { // */}}
	// {{define "distinctCollectRightOuter"}}
	for i := uint16(0); i < batchSize; i++ {
		// Index of keys and outputs in the hash table is calculated as ID - 1.
		prober.buildIdx[i] = prober.groupID[i] - 1
		prober.probeIdx[i] = _SEL_IND

		prober.buildNil[i] = prober.groupID[i] == 0
	}
	// {{end}}
	// {{/*
}

func _DISTINCT_COLLECT_NO_OUTER(_SEL_IND uint16) { // */}}
	// {{define "distinctCollectNoOuter"}}
	for i := uint16(0); i < batchSize; i++ {
		if prober.groupID[i] != 0 {
			// Index of keys and outputs in the hash table is calculated as ID - 1.
			prober.buildIdx[nResults] = prober.groupID[i] - 1
			prober.probeIdx[nResults] = _SEL_IND
			nResults++
		}
	}
	// {{end}}
	// {{/*
}

// */}}

// rehash takes a element of a key (tuple representing a row of equality
// column values) at a given column and computes a new hash by applying a
// transformation to the existing hash.
func (ht *hashTable) rehash(
	buckets []uint64, keyIdx int, t types.T, col ColVec, nKeys uint64, sel []uint16,
) {
	switch t {
	// {{range $hashType := .HashTemplate}}
	case _TYPES_T:
		keys := col._TemplateType()
		if sel != nil {
			_REHASH_BODY(sel[i])
		} else {
			_REHASH_BODY(i)
		}

	// {{end}}
	default:
		panic(fmt.Sprintf("unhandled type %d", t))
	}
}

// checkCol determines if the current key column in the groupID buckets matches
// the specified equality column key. If there is a match, then the key is added
// to differs. If the bucket has reached the end, the key is rejected. If any
// element in the key is null, then there is no match.
func (prober *hashJoinProber) checkCol(t types.T, keyColIdx int, nToCheck uint16, sel []uint16) {
	switch t {
	// {{range $neType := .NETemplate}}
	case _TYPES_T:
		buildVec := prober.ht.vals[prober.ht.keyCols[keyColIdx]]
		probeVec := prober.keys[keyColIdx]

		buildKeys := buildVec._TemplateType()
		probeKeys := probeVec._TemplateType()

		if sel != nil {
			_CHECK_COL_WITH_NULLS(sel[prober.toCheck[i]])
		} else {
			_CHECK_COL_WITH_NULLS(prober.toCheck[i])
		}
		// {{end}}
	default:
		panic(fmt.Sprintf("unhandled type %d", t))
	}
}

// collect prepares the buildIdx and probeIdx arrays where the buildIdx and
// probeIdx at each index are joined to make an output row. The total number of
// resulting rows is returned.
func (prober *hashJoinProber) collect(batch ColBatch, batchSize uint16, sel []uint16) uint16 {
	nResults := uint16(0)

	if prober.spec.outer {
		if sel != nil {
			_COLLECT_RIGHT_OUTER(sel[i])
		} else {
			_COLLECT_RIGHT_OUTER(i)
		}
	} else {
		if sel != nil {
			_COLLECT_NO_OUTER(sel[i])
		} else {
			_COLLECT_NO_OUTER(i)
		}
	}

	return nResults
}

// distinctCollect prepares the batch with the joined output columns where the build
// row index for each probe row is given in the groupID slice. This function
// requires assumes a N-1 hash join.
func (prober *hashJoinProber) distinctCollect(
	batch ColBatch, batchSize uint16, sel []uint16,
) uint16 {
	nResults := uint16(0)

	if prober.spec.outer {
		nResults = batchSize

		if sel != nil {
			_DISTINCT_COLLECT_RIGHT_OUTER(sel[i])
		} else {
			_DISTINCT_COLLECT_RIGHT_OUTER(i)
		}
	} else {
		if sel != nil {
			_DISTINCT_COLLECT_NO_OUTER(sel[i])
		} else {
			_DISTINCT_COLLECT_NO_OUTER(i)
		}
	}

	if prober.buildOuter {
		// if buildOuter is set, we need to mark all of the build table rows
		// that have been probed.
		for _, ID := range prober.groupID {
			prober.ht.visited[ID] = true
		}
	}

	return nResults
}

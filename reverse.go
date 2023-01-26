// Copyright 2023 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package caps

// Reverse returns a copy of the specified slice with the sequence of elements
// reversed.
func Reverse[S ~[]E, E any](s S) S {
	scopy := make(S, len(s))
	copy(scopy, s)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		scopy[i], scopy[j] = scopy[j], scopy[i]
	}
	return scopy
}

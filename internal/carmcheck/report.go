// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package carmcheck

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteText renders the report as human-readable text, one line per finding,
// grouped by check name.
func (r *Report) WriteText(w io.Writer) error {
	if len(r.Findings) == 0 {
		_, err := fmt.Fprintln(w, "no findings")
		return err
	}
	for _, f := range r.Findings {
		loc := f.Check
		if f.Namespace != "" {
			loc = fmt.Sprintf("%s[%s]", f.Check, f.Namespace)
		}
		if _, err := fmt.Fprintf(w, "[%s] %s: %s\n", strings.ToUpper(string(f.Severity)), loc, f.Message); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON renders the report as JSON.
func (r *Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

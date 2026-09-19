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
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// checkTeamID implements ADR-015 §5.8.2 rule 5 / §5.8.4 precondition 5
// (team-id half): services.k8s.aws/team-id on a kropath-managed namespace is
// unsupported. Under ACK's TeamLevelCARM feature gate the team branch
// resolves the role from the team map, and the owner-account-id branch this
// precondition depends on never executes — it is an "else if".
func checkTeamID(managed []corev1.Namespace) []Finding {
	var findings []Finding
	for _, ns := range managed {
		if v := ns.Annotations[TeamIDAnnotation]; v != "" {
			findings = append(findings, Finding{
				Check:     "team-id",
				Severity:  SeverityBlocking,
				Namespace: ns.Name,
				Message: fmt.Sprintf(
					"namespace %q carries %s=%q; kropath's computed account will not match what ACK resolves via the team map (ADR-015 §5.8.2 rule 5, TeamAnnotationUnsupported)",
					ns.Name, TeamIDAnnotation, v,
				),
			})
		}
	}
	if len(findings) == 0 {
		findings = append(findings, Finding{
			Check:    "team-id",
			Severity: SeverityOK,
			Message:  fmt.Sprintf("no kropath-managed namespace carries %s", TeamIDAnnotation),
		})
	}
	return findings
}

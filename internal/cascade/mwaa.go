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

package cascade

// MWAAKropathSection holds the MWAA-family governance fields from
// KropathConfig.spec.mandatory.mwaa / .defaults.mwaa (ADR-015 §3.5).
//
// Zero value for string fields is "" (not enforced).
// Zero value for int64 fields is 0 (not enforced).
// Boolean pointer fields use nil = not set (falls through); false = explicitly disabled.
//
// namingTemplate, syncedLabels, and syncedAnnotations are NOT present here;
// they live only in MWAAConfig (MWAAConfigSection).
// The Tags field is populated by the reconciler from KropathConfig.spec.mandatory.tags /
// .defaults.tags so that tag union merge flows through MergeMWAACascade.
type MWAAKropathSection struct {
	AirflowVersion               string            `json:"airflowVersion,omitempty"`
	EnvironmentClass             string            `json:"environmentClass,omitempty"`
	WebserverAccessMode          string            `json:"webserverAccessMode,omitempty"`
	EndpointManagement           string            `json:"endpointManagement,omitempty"`
	KmsKeyARN                    string            `json:"kmsKeyARN,omitempty"`
	DagProcessingLogsEnabled     *bool             `json:"dagProcessingLogsEnabled,omitempty"`
	DagProcessingLogsLevel       string            `json:"dagProcessingLogsLevel,omitempty"`
	SchedulerLogsEnabled         *bool             `json:"schedulerLogsEnabled,omitempty"`
	SchedulerLogsLevel           string            `json:"schedulerLogsLevel,omitempty"`
	TaskLogsEnabled              *bool             `json:"taskLogsEnabled,omitempty"`
	TaskLogsLevel                string            `json:"taskLogsLevel,omitempty"`
	WebserverLogsEnabled         *bool             `json:"webserverLogsEnabled,omitempty"`
	WebserverLogsLevel           string            `json:"webserverLogsLevel,omitempty"`
	WorkerLogsEnabled            *bool             `json:"workerLogsEnabled,omitempty"`
	WorkerLogsLevel              string            `json:"workerLogsLevel,omitempty"`
	MaxWorkers                   int64             `json:"maxWorkers,omitempty"`
	MinWorkers                   int64             `json:"minWorkers,omitempty"`
	MaxWebservers                int64             `json:"maxWebservers,omitempty"`
	MinWebservers                int64             `json:"minWebservers,omitempty"`
	Schedulers                   int64             `json:"schedulers,omitempty"`
	WeeklyMaintenanceWindowStart string            `json:"weeklyMaintenanceWindowStart,omitempty"`
	AirflowConfigurationOptions  map[string]string `json:"airflowConfigurationOptions,omitempty"`
	// Tags are tier-level cloud resource tags injected from KropathConfig.spec.mandatory.tags /
	// .defaults.tags by the reconciler before calling MergeMWAACascade.
	Tags map[string]string `json:"tags,omitempty"`
}

// MWAAConfigSection holds the MWAA governance fields from MWAAConfig.spec.mandatory
// or MWAAConfig.spec.defaults (per-type ResourceConfig, ADR-015 §3.5).
type MWAAConfigSection struct {
	AirflowVersion               string            `json:"airflowVersion,omitempty"`
	EnvironmentClass             string            `json:"environmentClass,omitempty"`
	WebserverAccessMode          string            `json:"webserverAccessMode,omitempty"`
	EndpointManagement           string            `json:"endpointManagement,omitempty"`
	KmsKeyARN                    string            `json:"kmsKeyARN,omitempty"`
	DagProcessingLogsEnabled     *bool             `json:"dagProcessingLogsEnabled,omitempty"`
	DagProcessingLogsLevel       string            `json:"dagProcessingLogsLevel,omitempty"`
	SchedulerLogsEnabled         *bool             `json:"schedulerLogsEnabled,omitempty"`
	SchedulerLogsLevel           string            `json:"schedulerLogsLevel,omitempty"`
	TaskLogsEnabled              *bool             `json:"taskLogsEnabled,omitempty"`
	TaskLogsLevel                string            `json:"taskLogsLevel,omitempty"`
	WebserverLogsEnabled         *bool             `json:"webserverLogsEnabled,omitempty"`
	WebserverLogsLevel           string            `json:"webserverLogsLevel,omitempty"`
	WorkerLogsEnabled            *bool             `json:"workerLogsEnabled,omitempty"`
	WorkerLogsLevel              string            `json:"workerLogsLevel,omitempty"`
	MaxWorkers                   int64             `json:"maxWorkers,omitempty"`
	MinWorkers                   int64             `json:"minWorkers,omitempty"`
	MaxWebservers                int64             `json:"maxWebservers,omitempty"`
	MinWebservers                int64             `json:"minWebservers,omitempty"`
	Schedulers                   int64             `json:"schedulers,omitempty"`
	WeeklyMaintenanceWindowStart string            `json:"weeklyMaintenanceWindowStart,omitempty"`
	AirflowConfigurationOptions  map[string]string `json:"airflowConfigurationOptions,omitempty"`
	// NamingTemplate is the environment naming template (e.g. "mwaa-{namespace}-{name}").
	// Governed only at MWAAConfig levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.mwaa does NOT carry namingTemplate.
	NamingTemplate    string            `json:"namingTemplate,omitempty"`
	SyncedLabels      map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// EffectiveMWAASection is one tier (mandatory or defaults) of the merged MWAA governance
// result written into MWAAConfig.status.effectiveConfig by the controller.
type EffectiveMWAASection struct {
	AirflowVersion               string            `json:"airflowVersion,omitempty"`
	EnvironmentClass             string            `json:"environmentClass,omitempty"`
	WebserverAccessMode          string            `json:"webserverAccessMode,omitempty"`
	EndpointManagement           string            `json:"endpointManagement,omitempty"`
	KmsKeyARN                    string            `json:"kmsKeyARN,omitempty"`
	DagProcessingLogsEnabled     *bool             `json:"dagProcessingLogsEnabled,omitempty"`
	DagProcessingLogsLevel       string            `json:"dagProcessingLogsLevel,omitempty"`
	SchedulerLogsEnabled         *bool             `json:"schedulerLogsEnabled,omitempty"`
	SchedulerLogsLevel           string            `json:"schedulerLogsLevel,omitempty"`
	TaskLogsEnabled              *bool             `json:"taskLogsEnabled,omitempty"`
	TaskLogsLevel                string            `json:"taskLogsLevel,omitempty"`
	WebserverLogsEnabled         *bool             `json:"webserverLogsEnabled,omitempty"`
	WebserverLogsLevel           string            `json:"webserverLogsLevel,omitempty"`
	WorkerLogsEnabled            *bool             `json:"workerLogsEnabled,omitempty"`
	WorkerLogsLevel              string            `json:"workerLogsLevel,omitempty"`
	MaxWorkers                   int64             `json:"maxWorkers,omitempty"`
	MinWorkers                   int64             `json:"minWorkers,omitempty"`
	MaxWebservers                int64             `json:"maxWebservers,omitempty"`
	MinWebservers                int64             `json:"minWebservers,omitempty"`
	Schedulers                   int64             `json:"schedulers,omitempty"`
	WeeklyMaintenanceWindowStart string            `json:"weeklyMaintenanceWindowStart,omitempty"`
	AirflowConfigurationOptions  map[string]string `json:"airflowConfigurationOptions,omitempty"`
	NamingTemplate               string            `json:"namingTemplate,omitempty"`
	SyncedLabels                 map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations            map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                         map[string]string `json:"tags,omitempty"`
}

// EffectiveMWAAConfig is the merged MWAA governance result written into
// MWAAConfig.status.effectiveConfig by the controller.
type EffectiveMWAAConfig struct {
	Mandatory EffectiveMWAASection `json:"mandatory"`
	Defaults  EffectiveMWAASection `json:"defaults"`
}

// MergeMWAACascade merges MWAA governance fields from all cascade sources and
// returns the effective configuration to be written to status.effectiveConfig.
//
// Nine-level priority chain for MWAA (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.mwaa)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.mwaa)
//	Level 3 — globalMWAACfgMandatory  (MWAAConfig in kro-system, mandatory)
//	Level 4 — localMWAACfgMandatory   (MWAAConfig in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localMWAACfgDefaults    (MWAAConfig in resource namespace, defaults)
//	Level 7 — globalMWAACfgDefaults   (MWAAConfig in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.mwaa)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.mwaa)
//
// Scalar string merge: firstNonEmptyString in priority order (lowest number wins).
// int64 merge: firstNonZeroInt64 in priority order.
// *bool pointer merge: firstNonNilBoolPtr in priority order — nil = not set (falls through).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from MWAAConfig levels only (3-4 mandatory, 6-7 defaults).
// NamingTemplate: governed only at MWAAConfig levels (3-4 mandatory, 6-7 defaults).
// AirflowConfigurationOptions: key-priority map merge (same priority as tags for each tier).
func MergeMWAACascade(
	globalKropathMandatory MWAAKropathSection, // level 1
	localKropathMandatory MWAAKropathSection,  // level 2
	globalMWAACfgMandatory MWAAConfigSection,  // level 3
	localMWAACfgMandatory MWAAConfigSection,   // level 4
	localMWAACfgDefaults MWAAConfigSection,    // level 6
	globalMWAACfgDefaults MWAAConfigSection,   // level 7
	localKropathDefaults MWAAKropathSection,   // level 8
	globalKropathDefaults MWAAKropathSection,  // level 9
) EffectiveMWAAConfig {
	return EffectiveMWAAConfig{
		Mandatory: EffectiveMWAASection{
			AirflowVersion: firstNonEmptyString(
				globalKropathMandatory.AirflowVersion,
				localKropathMandatory.AirflowVersion,
				globalMWAACfgMandatory.AirflowVersion,
				localMWAACfgMandatory.AirflowVersion,
			),
			EnvironmentClass: firstNonEmptyString(
				globalKropathMandatory.EnvironmentClass,
				localKropathMandatory.EnvironmentClass,
				globalMWAACfgMandatory.EnvironmentClass,
				localMWAACfgMandatory.EnvironmentClass,
			),
			WebserverAccessMode: firstNonEmptyString(
				globalKropathMandatory.WebserverAccessMode,
				localKropathMandatory.WebserverAccessMode,
				globalMWAACfgMandatory.WebserverAccessMode,
				localMWAACfgMandatory.WebserverAccessMode,
			),
			EndpointManagement: firstNonEmptyString(
				globalKropathMandatory.EndpointManagement,
				localKropathMandatory.EndpointManagement,
				globalMWAACfgMandatory.EndpointManagement,
				localMWAACfgMandatory.EndpointManagement,
			),
			KmsKeyARN: firstNonEmptyString(
				globalKropathMandatory.KmsKeyARN,
				localKropathMandatory.KmsKeyARN,
				globalMWAACfgMandatory.KmsKeyARN,
				localMWAACfgMandatory.KmsKeyARN,
			),
			DagProcessingLogsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.DagProcessingLogsEnabled,
				localKropathMandatory.DagProcessingLogsEnabled,
				globalMWAACfgMandatory.DagProcessingLogsEnabled,
				localMWAACfgMandatory.DagProcessingLogsEnabled,
			),
			DagProcessingLogsLevel: firstNonEmptyString(
				globalKropathMandatory.DagProcessingLogsLevel,
				localKropathMandatory.DagProcessingLogsLevel,
				globalMWAACfgMandatory.DagProcessingLogsLevel,
				localMWAACfgMandatory.DagProcessingLogsLevel,
			),
			SchedulerLogsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.SchedulerLogsEnabled,
				localKropathMandatory.SchedulerLogsEnabled,
				globalMWAACfgMandatory.SchedulerLogsEnabled,
				localMWAACfgMandatory.SchedulerLogsEnabled,
			),
			SchedulerLogsLevel: firstNonEmptyString(
				globalKropathMandatory.SchedulerLogsLevel,
				localKropathMandatory.SchedulerLogsLevel,
				globalMWAACfgMandatory.SchedulerLogsLevel,
				localMWAACfgMandatory.SchedulerLogsLevel,
			),
			TaskLogsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.TaskLogsEnabled,
				localKropathMandatory.TaskLogsEnabled,
				globalMWAACfgMandatory.TaskLogsEnabled,
				localMWAACfgMandatory.TaskLogsEnabled,
			),
			TaskLogsLevel: firstNonEmptyString(
				globalKropathMandatory.TaskLogsLevel,
				localKropathMandatory.TaskLogsLevel,
				globalMWAACfgMandatory.TaskLogsLevel,
				localMWAACfgMandatory.TaskLogsLevel,
			),
			WebserverLogsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.WebserverLogsEnabled,
				localKropathMandatory.WebserverLogsEnabled,
				globalMWAACfgMandatory.WebserverLogsEnabled,
				localMWAACfgMandatory.WebserverLogsEnabled,
			),
			WebserverLogsLevel: firstNonEmptyString(
				globalKropathMandatory.WebserverLogsLevel,
				localKropathMandatory.WebserverLogsLevel,
				globalMWAACfgMandatory.WebserverLogsLevel,
				localMWAACfgMandatory.WebserverLogsLevel,
			),
			WorkerLogsEnabled: firstNonNilBoolPtr(
				globalKropathMandatory.WorkerLogsEnabled,
				localKropathMandatory.WorkerLogsEnabled,
				globalMWAACfgMandatory.WorkerLogsEnabled,
				localMWAACfgMandatory.WorkerLogsEnabled,
			),
			WorkerLogsLevel: firstNonEmptyString(
				globalKropathMandatory.WorkerLogsLevel,
				localKropathMandatory.WorkerLogsLevel,
				globalMWAACfgMandatory.WorkerLogsLevel,
				localMWAACfgMandatory.WorkerLogsLevel,
			),
			MaxWorkers: firstNonZeroInt64(
				globalKropathMandatory.MaxWorkers,
				localKropathMandatory.MaxWorkers,
				globalMWAACfgMandatory.MaxWorkers,
				localMWAACfgMandatory.MaxWorkers,
			),
			MinWorkers: firstNonZeroInt64(
				globalKropathMandatory.MinWorkers,
				localKropathMandatory.MinWorkers,
				globalMWAACfgMandatory.MinWorkers,
				localMWAACfgMandatory.MinWorkers,
			),
			MaxWebservers: firstNonZeroInt64(
				globalKropathMandatory.MaxWebservers,
				localKropathMandatory.MaxWebservers,
				globalMWAACfgMandatory.MaxWebservers,
				localMWAACfgMandatory.MaxWebservers,
			),
			MinWebservers: firstNonZeroInt64(
				globalKropathMandatory.MinWebservers,
				localKropathMandatory.MinWebservers,
				globalMWAACfgMandatory.MinWebservers,
				localMWAACfgMandatory.MinWebservers,
			),
			Schedulers: firstNonZeroInt64(
				globalKropathMandatory.Schedulers,
				localKropathMandatory.Schedulers,
				globalMWAACfgMandatory.Schedulers,
				localMWAACfgMandatory.Schedulers,
			),
			WeeklyMaintenanceWindowStart: firstNonEmptyString(
				globalKropathMandatory.WeeklyMaintenanceWindowStart,
				localKropathMandatory.WeeklyMaintenanceWindowStart,
				globalMWAACfgMandatory.WeeklyMaintenanceWindowStart,
				localMWAACfgMandatory.WeeklyMaintenanceWindowStart,
			),
			// AirflowConfigurationOptions: key-priority merge; L4 added first (lowest), L1 wins on conflict.
			AirflowConfigurationOptions: mergeMaps(
				localMWAACfgMandatory.AirflowConfigurationOptions,
				globalMWAACfgMandatory.AirflowConfigurationOptions,
				localKropathMandatory.AirflowConfigurationOptions,
				globalKropathMandatory.AirflowConfigurationOptions,
			),
			// NamingTemplate: MWAAConfig levels only (3, 4); KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalMWAACfgMandatory.NamingTemplate,
				localMWAACfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from MWAAConfig levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localMWAACfgMandatory.SyncedLabels,
				globalMWAACfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localMWAACfgMandatory.SyncedAnnotations,
				globalMWAACfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localMWAACfgMandatory.Tags,
				globalMWAACfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveMWAASection{
			AirflowVersion: firstNonEmptyString(
				localMWAACfgDefaults.AirflowVersion,
				globalMWAACfgDefaults.AirflowVersion,
				localKropathDefaults.AirflowVersion,
				globalKropathDefaults.AirflowVersion,
			),
			EnvironmentClass: firstNonEmptyString(
				localMWAACfgDefaults.EnvironmentClass,
				globalMWAACfgDefaults.EnvironmentClass,
				localKropathDefaults.EnvironmentClass,
				globalKropathDefaults.EnvironmentClass,
			),
			WebserverAccessMode: firstNonEmptyString(
				localMWAACfgDefaults.WebserverAccessMode,
				globalMWAACfgDefaults.WebserverAccessMode,
				localKropathDefaults.WebserverAccessMode,
				globalKropathDefaults.WebserverAccessMode,
			),
			EndpointManagement: firstNonEmptyString(
				localMWAACfgDefaults.EndpointManagement,
				globalMWAACfgDefaults.EndpointManagement,
				localKropathDefaults.EndpointManagement,
				globalKropathDefaults.EndpointManagement,
			),
			KmsKeyARN: firstNonEmptyString(
				localMWAACfgDefaults.KmsKeyARN,
				globalMWAACfgDefaults.KmsKeyARN,
				localKropathDefaults.KmsKeyARN,
				globalKropathDefaults.KmsKeyARN,
			),
			DagProcessingLogsEnabled: firstNonNilBoolPtr(
				localMWAACfgDefaults.DagProcessingLogsEnabled,
				globalMWAACfgDefaults.DagProcessingLogsEnabled,
				localKropathDefaults.DagProcessingLogsEnabled,
				globalKropathDefaults.DagProcessingLogsEnabled,
			),
			DagProcessingLogsLevel: firstNonEmptyString(
				localMWAACfgDefaults.DagProcessingLogsLevel,
				globalMWAACfgDefaults.DagProcessingLogsLevel,
				localKropathDefaults.DagProcessingLogsLevel,
				globalKropathDefaults.DagProcessingLogsLevel,
			),
			SchedulerLogsEnabled: firstNonNilBoolPtr(
				localMWAACfgDefaults.SchedulerLogsEnabled,
				globalMWAACfgDefaults.SchedulerLogsEnabled,
				localKropathDefaults.SchedulerLogsEnabled,
				globalKropathDefaults.SchedulerLogsEnabled,
			),
			SchedulerLogsLevel: firstNonEmptyString(
				localMWAACfgDefaults.SchedulerLogsLevel,
				globalMWAACfgDefaults.SchedulerLogsLevel,
				localKropathDefaults.SchedulerLogsLevel,
				globalKropathDefaults.SchedulerLogsLevel,
			),
			TaskLogsEnabled: firstNonNilBoolPtr(
				localMWAACfgDefaults.TaskLogsEnabled,
				globalMWAACfgDefaults.TaskLogsEnabled,
				localKropathDefaults.TaskLogsEnabled,
				globalKropathDefaults.TaskLogsEnabled,
			),
			TaskLogsLevel: firstNonEmptyString(
				localMWAACfgDefaults.TaskLogsLevel,
				globalMWAACfgDefaults.TaskLogsLevel,
				localKropathDefaults.TaskLogsLevel,
				globalKropathDefaults.TaskLogsLevel,
			),
			WebserverLogsEnabled: firstNonNilBoolPtr(
				localMWAACfgDefaults.WebserverLogsEnabled,
				globalMWAACfgDefaults.WebserverLogsEnabled,
				localKropathDefaults.WebserverLogsEnabled,
				globalKropathDefaults.WebserverLogsEnabled,
			),
			WebserverLogsLevel: firstNonEmptyString(
				localMWAACfgDefaults.WebserverLogsLevel,
				globalMWAACfgDefaults.WebserverLogsLevel,
				localKropathDefaults.WebserverLogsLevel,
				globalKropathDefaults.WebserverLogsLevel,
			),
			WorkerLogsEnabled: firstNonNilBoolPtr(
				localMWAACfgDefaults.WorkerLogsEnabled,
				globalMWAACfgDefaults.WorkerLogsEnabled,
				localKropathDefaults.WorkerLogsEnabled,
				globalKropathDefaults.WorkerLogsEnabled,
			),
			WorkerLogsLevel: firstNonEmptyString(
				localMWAACfgDefaults.WorkerLogsLevel,
				globalMWAACfgDefaults.WorkerLogsLevel,
				localKropathDefaults.WorkerLogsLevel,
				globalKropathDefaults.WorkerLogsLevel,
			),
			MaxWorkers: firstNonZeroInt64(
				localMWAACfgDefaults.MaxWorkers,
				globalMWAACfgDefaults.MaxWorkers,
				localKropathDefaults.MaxWorkers,
				globalKropathDefaults.MaxWorkers,
			),
			MinWorkers: firstNonZeroInt64(
				localMWAACfgDefaults.MinWorkers,
				globalMWAACfgDefaults.MinWorkers,
				localKropathDefaults.MinWorkers,
				globalKropathDefaults.MinWorkers,
			),
			MaxWebservers: firstNonZeroInt64(
				localMWAACfgDefaults.MaxWebservers,
				globalMWAACfgDefaults.MaxWebservers,
				localKropathDefaults.MaxWebservers,
				globalKropathDefaults.MaxWebservers,
			),
			MinWebservers: firstNonZeroInt64(
				localMWAACfgDefaults.MinWebservers,
				globalMWAACfgDefaults.MinWebservers,
				localKropathDefaults.MinWebservers,
				globalKropathDefaults.MinWebservers,
			),
			Schedulers: firstNonZeroInt64(
				localMWAACfgDefaults.Schedulers,
				globalMWAACfgDefaults.Schedulers,
				localKropathDefaults.Schedulers,
				globalKropathDefaults.Schedulers,
			),
			WeeklyMaintenanceWindowStart: firstNonEmptyString(
				localMWAACfgDefaults.WeeklyMaintenanceWindowStart,
				globalMWAACfgDefaults.WeeklyMaintenanceWindowStart,
				localKropathDefaults.WeeklyMaintenanceWindowStart,
				globalKropathDefaults.WeeklyMaintenanceWindowStart,
			),
			// AirflowConfigurationOptions: key-priority merge; L9 added first (lowest), L6 wins on conflict.
			AirflowConfigurationOptions: mergeMaps(
				globalKropathDefaults.AirflowConfigurationOptions,
				localKropathDefaults.AirflowConfigurationOptions,
				globalMWAACfgDefaults.AirflowConfigurationOptions,
				localMWAACfgDefaults.AirflowConfigurationOptions,
			),
			// NamingTemplate: MWAAConfig levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localMWAACfgDefaults.NamingTemplate,
				globalMWAACfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from MWAAConfig levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalMWAACfgDefaults.SyncedLabels,
				localMWAACfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalMWAACfgDefaults.SyncedAnnotations,
				localMWAACfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalMWAACfgDefaults.Tags,
				localMWAACfgDefaults.Tags,
			),
		},
	}
}

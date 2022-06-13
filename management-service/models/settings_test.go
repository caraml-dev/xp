package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/turing-experiments/common/api/schema"
	_pubsub "github.com/gojek/turing-experiments/common/pubsub"
	tu "github.com/gojek/turing-experiments/management-service/internal/testutils"
)

var testExperimentationConfig = ExperimentationConfig{
	Segmenters: ProjectSegmenters{
		Names: []string{"seg1", "seg2"},
		Variables: map[string][]string{
			"seg1": {"exp-var-1", "exp-var-2"},
			"seg2": {"exp-var-3"},
		},
	},
	RandomizationKey:      "rkey",
	S2IDClusteringEnabled: true,
}

func TestExperimentationConfigValue(t *testing.T) {
	value, err := testExperimentationConfig.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		{
			"segmenters": {
				"names": ["seg1","seg2"],
				"variables": {
				 	"seg1": ["exp-var-1","exp-var-2"],
					"seg2": ["exp-var-3"]
				 }
			 },
			"randomization_key": "rkey",
			"enable_s2id_clustering": true
		}
	`, string(byteValue))
}

func TestExperimentationConfigScan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		errString string
		expected  ExperimentationConfig
	}{
		{
			name: "success",
			value: []byte(`
				{
					"segmenters": {
						"names": ["seg1","seg2"],
						"variables": {
							"seg1": ["exp-var-1","exp-var-2"],
							"seg2": ["exp-var-3"]
						 }
					 },
					"randomization_key": "rkey",
					"enable_s2id_clustering": true
				}
			`),
			expected: testExperimentationConfig,
		},
		{
			name:      "failure | invalid value",
			value:     100,
			errString: "type assertion to []byte failed",
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			var expCfg ExperimentationConfig
			err := expCfg.Scan(data.value)
			if data.errString == "" {
				// Success
				require.NoError(t, err)
				assert.Equal(t, data.expected, expCfg)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestSettingsToApiSchema(t *testing.T) {
	tests := []struct {
		Name     string
		Settings Settings
		Expected schema.ProjectSettings
	}{
		{
			Name: "No experiment validation URL",
			Settings: Settings{
				Model: Model{
					CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
					UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				},
				ProjectID: ID(1),
				Username:  "client-1",
				Passkey:   "passkey-1",
				Config: &ExperimentationConfig{
					Segmenters: ProjectSegmenters{
						Names: []string{"seg3", "seg4"},
						Variables: map[string][]string{
							"seg3": {"exp_var_3.1", "exp_var_3.2"},
							"seg4": {"exp_var_4"},
						},
					},
					RandomizationKey: "rand",
				},
			},
			Expected: schema.ProjectSettings{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				ProjectId: 1,
				Username:  "client-1",
				Passkey:   "passkey-1",
				Segmenters: schema.ProjectSegmenters{
					Names: []string{"seg3", "seg4"},
					Variables: schema.ProjectSegmenters_Variables{
						AdditionalProperties: map[string][]string{
							"seg3": {"exp_var_3.1", "exp_var_3.2"},
							"seg4": {"exp_var_4"},
						}},
				},
				RandomizationKey:     "rand",
				EnableS2idClustering: false,
			},
		},
		{
			Name: "Experiment validation URL set",
			Settings: Settings{
				Model: Model{
					CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
					UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				},
				ProjectID: ID(2),
				Username:  "client-2",
				Passkey:   "passkey-2",
				Config: &ExperimentationConfig{
					Segmenters: ProjectSegmenters{
						Names: []string{"seg5", "seg6"},
						Variables: map[string][]string{
							"seg5": {"exp_var_5.1", "exp_var_5.2"},
							"seg6": {"exp_var_6"},
						},
					},
					RandomizationKey:      "rand-2",
					S2IDClusteringEnabled: true,
				},
			},
			Expected: schema.ProjectSettings{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				ProjectId: 2,
				Username:  "client-2",
				Passkey:   "passkey-2",
				Segmenters: schema.ProjectSegmenters{
					Names: []string{"seg5", "seg6"},
					Variables: schema.ProjectSegmenters_Variables{
						AdditionalProperties: map[string][]string{
							"seg5": {"exp_var_5.1", "exp_var_5.2"},
							"seg6": {"exp_var_6"},
						}},
				},
				RandomizationKey:     "rand-2",
				EnableS2idClustering: true,
			},
		},
		{
			Name: "Treatment validation set",
			Settings: Settings{
				Model: Model{
					CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
					UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				},
				ProjectID: ID(3),
				Username:  "client-3",
				Passkey:   "passkey-3",
				Config: &ExperimentationConfig{
					Segmenters: ProjectSegmenters{
						Names: []string{"seg5", "seg6"},
						Variables: map[string][]string{
							"seg5": {"exp_var_5.1", "exp_var_5.2"},
							"seg6": {"exp_var_6"},
						},
					},
					RandomizationKey:      "rand-3",
					S2IDClusteringEnabled: false,
				},
				TreatmentSchema: &TreatmentSchema{
					Rules: []Rule{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_2",
							Predicate: "predicate_2",
						},
					},
				},
				ValidationUrl: nil,
			},
			Expected: schema.ProjectSettings{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				ProjectId: 3,
				Username:  "client-3",
				Passkey:   "passkey-3",
				Segmenters: schema.ProjectSegmenters{
					Names: []string{"seg5", "seg6"},
					Variables: schema.ProjectSegmenters_Variables{
						AdditionalProperties: map[string][]string{
							"seg5": {"exp_var_5.1", "exp_var_5.2"},
							"seg6": {"exp_var_6"},
						}},
				},
				TreatmentSchema: &schema.TreatmentSchema{
					Rules: schema.Rules{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_2",
							Predicate: "predicate_2",
						},
					},
				},
				ValidationUrl:        nil,
				RandomizationKey:     "rand-3",
				EnableS2idClustering: false,
			},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			tu.AssertEqualValues(t, data.Expected, data.Settings.ToApiSchema())
		})
	}
}

func TestSettingsToProtoSchema(t *testing.T) {
	createdUpdatedAt := time.Now()
	projectId := int64(1)
	randomizationKey := "random"
	username := "user1"
	passkey := "pass1"
	testSettings := Settings{
		Model: Model{
			CreatedAt: createdUpdatedAt,
			UpdatedAt: createdUpdatedAt,
		},
		ProjectID: ID(projectId),
		Username:  username,
		Passkey:   passkey,
		Config: &ExperimentationConfig{
			Segmenters: ProjectSegmenters{
				Names: []string{"seg1"},
				Variables: map[string][]string{
					"seg1": {"exp-var-1", "exp-var-2"},
				},
			},
			RandomizationKey:      randomizationKey,
			S2IDClusteringEnabled: false,
		},
	}

	pubSubExperimentVariables := _pubsub.ExperimentVariables{
		Value: []string{"exp-var-1", "exp-var-2"},
	}

	pubSubSegmenters := _pubsub.Segmenters{
		Names: []string{"seg1"},
		Variables: map[string]*_pubsub.ExperimentVariables{
			"seg1": &pubSubExperimentVariables,
		},
	}
	assert.Equal(t, _pubsub.ProjectSettings{
		ProjectId:        projectId,
		CreatedAt:        timestamppb.New(createdUpdatedAt),
		RandomizationKey: randomizationKey,
		Segmenters:       &pubSubSegmenters,
		UpdatedAt:        timestamppb.New(createdUpdatedAt),
		Username:         username,
		Passkey:          passkey,
	}, testSettings.ToProtoSchema())
}

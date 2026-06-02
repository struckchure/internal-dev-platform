package types

import (
	"fmt"
	"strings"

	storm "github.com/Overal-X/formatio.storm"
	"gopkg.in/yaml.v2"
)

const defaultJobRunsOn = "self-hosted"

// ActionJobs is a []storm.Job that accepts Storm, GitHub Actions, and legacy list-of-maps YAML.
type ActionJobs []storm.Job

func (j *ActionJobs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	jobs, err := parseJobsRaw(raw)
	if err != nil {
		return err
	}
	*j = ActionJobs(jobs)
	return nil
}

// ParseAction unmarshals a workflow action file into Action.
func ParseAction(content string) (*Action, error) {
	var config Action
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}
	normalizeActionJobs(&config)
	return &config, nil
}

func normalizeActionJobs(config *Action) {
	jobs := []storm.Job(config.Jobs)
	for i := range jobs {
		if strings.TrimSpace(jobs[i].RunsOn) == "" {
			jobs[i].RunsOn = defaultJobRunsOn
		}
	}
	config.Jobs = ActionJobs(jobs)
}

func parseJobsRaw(raw interface{}) ([]storm.Job, error) {
	if raw == nil {
		return nil, nil
	}

	switch v := raw.(type) {
	case []interface{}:
		return parseJobsSequence(v)
	case map[interface{}]interface{}:
		return parseJobsMap(toStringKeyedMap(v))
	case map[string]interface{}:
		return parseJobsMap(v)
	default:
		return nil, fmt.Errorf("unsupported jobs type %T", raw)
	}
}

func parseJobsSequence(items []interface{}) ([]storm.Job, error) {
	jobs := make([]storm.Job, 0, len(items))
	for _, item := range items {
		job, err := parseJobItem(item)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func parseJobsMap(items map[string]interface{}) ([]storm.Job, error) {
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	// Stable order for maps with multiple jobs.
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}

	jobs := make([]storm.Job, 0, len(names))
	for _, name := range names {
		job, err := unmarshalJob(name, items[name])
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func parseJobItem(item interface{}) (storm.Job, error) {
	switch v := item.(type) {
	case map[interface{}]interface{}:
		m := toStringKeyedMap(v)
		if len(m) == 1 {
			if _, hasName := m["name"]; !hasName {
				for name, body := range m {
					return unmarshalJob(name, body)
				}
			}
		}
		return unmarshalJob("", m)
	case map[string]interface{}:
		if len(v) == 1 {
			if _, hasName := v["name"]; !hasName {
				for name, body := range v {
					return unmarshalJob(name, body)
				}
			}
		}
		return unmarshalJob("", v)
	default:
		return storm.Job{}, fmt.Errorf("unsupported job item type %T", item)
	}
}

func unmarshalJob(defaultName string, body interface{}) (storm.Job, error) {
	data, err := yaml.Marshal(body)
	if err != nil {
		return storm.Job{}, err
	}

	var job storm.Job
	if err := yaml.Unmarshal(data, &job); err != nil {
		return storm.Job{}, err
	}

	if strings.TrimSpace(job.Name) == "" {
		job.Name = defaultName
	}
	if strings.TrimSpace(job.RunsOn) == "" {
		job.RunsOn = defaultJobRunsOn
	}
	return job, nil
}

func toStringKeyedMap(m map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[fmt.Sprint(k)] = v
	}
	return out
}

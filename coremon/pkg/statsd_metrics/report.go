package statsd_metrics

import "fmt"

// Report should be used to send any data to the agent.
func Report(reportFn func(s Statter, tagSpec []string), tags ...Tags) {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil {
		return
	}

	reportFn(client, JoinTags(tags...))
}

type Tags map[string]string

// With allows to concatenate tagset together with a new tag pair.
func (t Tags) With(k, v string) Tags {
	if t == nil || len(t) == 0 {
		return map[string]string{
			k: v,
		}
	}

	t[k] = v
	return t
}

// WithBaseTags allows to inject metrics BaseTags into custom tag set. Useful when
// tags are used outside of StatsD reporting. For example, within InfluxDB points.
func (t Tags) WithBaseTags() Tags {
	baseTags := config.BaseTagsMap()
	allTags := make(Tags, len(baseTags)+len(t))

	for k, v := range baseTags {
		allTags[k] = v
	}

	for k, v := range t {
		allTags[k] = v
	}

	return allTags
}

// JoinTags joins the tags for the receiving agent.
func JoinTags(tags ...Tags) []string {
	if len(tags) == 0 {
		return []string{}
	}

	allTags := make([]string, 0, len(tags))
	for _, tagSet := range tags {
		for key, value := range tagSet {
			allTags = append(allTags, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return allTags
}

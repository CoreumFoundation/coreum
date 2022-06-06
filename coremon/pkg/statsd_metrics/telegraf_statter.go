package statsd_metrics

import (
	"strings"
	"time"

	statsd "github.com/alexcesaro/statsd"
)

type telegrafStatter struct {
	client *statsd.Client
	Statter
}

// telegrafStatter is wrapper of StatsD client to Telegraf, implementing the statter.
func newTelegrafStatter(opts ...statsd.Option) (Statter, error) {
	c, err := statsd.New(opts...)
	if err != nil {
		return nil, err
	}
	statter := &telegrafStatter{
		client: c,
	}
	return statter, nil
}

func (t *telegrafStatter) Count(bucket string, value interface{}, tags []string) (err error) {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Count(s, value)
	return nil
}

func (t *telegrafStatter) Incr(bucket string, tags []string) error {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Increment(s)
	return nil
}

func (t *telegrafStatter) Decr(bucket string, tags []string) error {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Count(s, -1)
	return nil
}

func (t *telegrafStatter) Gauge(bucket string, value interface{}, tags []string) error {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Gauge(s, value)
	return nil
}

func (t *telegrafStatter) Timing(bucket string, value time.Duration, tags []string) error {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Timing(s, int(value/time.Millisecond))
	return nil
}

func (t *telegrafStatter) Histogram(bucket string, value interface{}, tags []string) error {
	s := bucket + "," + strings.Join(tags, ",")
	t.client.Histogram(s, value)
	return nil
}

func (t *telegrafStatter) Close() error {
	t.client.Close()
	return nil
}

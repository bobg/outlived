package site

import (
	"container/list"
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

type taskService interface {
	queueEmpty(ctx context.Context, queue string) (bool, error)
	enqueueTask(ctx context.Context, queue, taskName, url string) error
}

type gCloudTasks cloudtasks.Client

func newGCloudTasks(ctx context.Context, options ...option.ClientOption) (*gCloudTasks, error) {
	client, err := cloudtasks.NewClient(ctx, options...)
	return (*gCloudTasks)(client), err
}

func (t *gCloudTasks) queueEmpty(ctx context.Context, queue string) (bool, error) {
	ltreq := &taskspb.ListTasksRequest{Parent: queue}
	iter := (*cloudtasks.Client)(t).ListTasks(ctx, ltreq)
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		return false, errors.Wrapf(err, "gCloudTasks: checking queue %s for emptiness", queue)
	}
	return err == iterator.Done, nil
}

func (t *gCloudTasks) enqueueTask(ctx context.Context, queue, taskName, url string) error {
	_, err := (*cloudtasks.Client)(t).CreateTask(ctx, &taskspb.CreateTaskRequest{
		Parent: queue,
		Task: &taskspb.Task{
			Name: taskName,
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_GET,
					RelativeUri: url,
				},
			},
		},
	})
	return errors.Wrapf(err, "gCloudTasks: enqueueing task %s, queue %s, url %s", taskName, queue, url)
}

type localTasks struct {
	base   *url.URL
	ctx    context.Context
	queues map[string]*list.List

	mu sync.Mutex // protects queues
}

func newLocalTasks(ctx context.Context, host string) *localTasks {
	return &localTasks{
		base: &url.URL{
			Scheme: "http",
			Host:   host,
		},
		ctx:    ctx,
		queues: make(map[string]*list.List),
	}
}

func (t *localTasks) queueEmpty(ctx context.Context, name string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	queue := t.queues[name]
	return queue == nil || queue.Len() == 0, nil
}

func (t *localTasks) enqueueTask(ctx context.Context, queueName, _, url string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	queue := t.queues[queueName]
	if queue == nil {
		queue = list.New()
		t.queues[queueName] = queue
		go t.process(queueName)
	}

	queue.PushBack(url)
	return nil
}

func (t *localTasks) process(queueName string) {
	log.Printf("starting queue processor for %s", queueName)
	defer log.Printf("exiting queue processor for %s", queueName)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return

		case <-ticker.C:
			urlstr := t.getTask(queueName)
			if urlstr == "" {
				continue
			}

			u, err := url.Parse(urlstr)
			if err != nil {
				log.Printf("localTasks, queue %s: parsing url %s: %s", queueName, urlstr, err)
				continue
			}

			_, err = http.Get(t.base.ResolveReference(u).String())
			if err != nil {
				log.Printf("localTasks, queue %s: during GET %s: %s", queueName, urlstr, err)
				continue
			}
		}
	}
}

func (t *localTasks) getTask(queueName string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	queue := t.queues[queueName]
	if queue == nil {
		return ""
	}
	el := queue.Front()
	if el == nil {
		return ""
	}
	return queue.Remove(el).(string)
}

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	filterpb "go.temporal.io/api/filter/v1"
	v110 "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/primitives/timestamp"
)

func main() {
	ctx := context.Background()
	addr, closeFn, err := GetKubectlPortForwardAddr(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("kubectl port-forward failed: %w", err)
	}

	cl, err := NewTemporalClient(ctx, addr)
	if err != nil {
		return fmt.Errorf("creating temporal client failed: %w", err)
	}
	defer closeFn()

	root := New()
	root.ParseAndRun(context.Background(), os.Args[1:])
}

func New() *ffcli.Command {
	listCmd := &ffcli.Command{
		Name:        "list",
		ShortUsage:  "pnscli job list",
		Exec:        nil,
	}

	historyCmd := &ffcli.Command{
		Name:       "history",
		ShortUsage: "pnscli job history <JOB ID>",
		ShortHelp:  "exports job execution history in csv format for further analysis",
		Exec: func(ctx context.Context, args []string) error {
			// TODO use same port forwarding
			return nil
		},
	}

	return &ffcli.Command{
		Name:        "job",
		ShortUsage:  "pnscli job <subcommand> [args]",
		Subcommands: []*ffcli.Command{listCmd, historyCmd},
	}
}



type temportalClient struct {
	Client client.Client
}

func NewTemporalClient(addr string) (*client.Client,  error) {
	cl, err := temporalClient(addr)
	if err != nil {
		return nil, fmt.Errorf("start Temporal client: %w", err)
	}

	return &cl, nil
}

func GetKubectlPortForwardAddr(ctx context.Context) (string, func(), error) {
	kctlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", nil, fmt.Errorf("kubectl not found: %w", err)
	}
	if err != nil {
		return "", nil, fmt.Errorf("kubectl not found: %w", err)
	}

	localPort, err := getFreeLocalPort()
	if err != nil {
		return "", nil, fmt.Errorf("no local port available: %w", err)
	}

	kubectlPortForward := &exec.Cmd{
		Path:   kctlPath,
		Args:   []string{"kubectl", "port-forward", "--cluster=ams4-1", "--namespace=pns-push", "service/temporal-frontend", localPort + ":7233"},
		Stderr: os.Stderr,
	}

	if err := kubectlPortForward.Start(); err != nil {
		return "", nil, err
	}
	// nolint: errcheck

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		// nolint: errcheck
		kubectlPortForward.Wait()
		cancel()
	}()

	// check port-forward is ready
	addr := net.JoinHostPort("127.0.0.1", localPort)
	timeout := time.Second
	// wait for port to open
PORT:
	for {
		select {
		case <-ctx.Done():
			return "", nil, fmt.Errorf("timeout")
		case <-time.After(timeout):
			log.Print("unable to connect to kubectl port-forward")
		default:
			// check port-forward is ready
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			conn.Close()
			break PORT
		}
	}

	return addr, func() {
		cancel()
		kubectlPortForward.Process.Signal(syscall.SIGTERM)
	}, nil

}

func (c temportalClient) list(ctx context.Context, args []string) error {
	printHeader()

	if err := printCommand(ctx, listOpened(c.Client), ">"); err != nil {
		return fmt.Errorf("print opened jobs: %w", err)
	}

	if err := printCommand(ctx, listCLosed(c.Client), ""); err != nil {
		return fmt.Errorf("print closed jobs: %w", err)
	}

	return nil
}

func listCLosed(client client.Client) commandFn {
	return func(ctx context.Context, pageToken []byte) (Responser, error) {
		return client.ListOpenWorkflow(ctx,
			&workflowservice.ListOpenWorkflowExecutionsRequest{
				Namespace:       "default",
				MaximumPageSize: 20,
				NextPageToken:   pageToken,
				StartTimeFilter: &filterpb.StartTimeFilter{
					LatestTime: timestamp.TimePtr(time.Now().UTC()),
				},
			})
	}

}

func listOpened(client client.Client) commandFn {
	return func(ctx context.Context, pageToken []byte) (Responser, error) {
		return client.ListOpenWorkflow(ctx,
			&workflowservice.ListOpenWorkflowExecutionsRequest{
				Namespace:       "default",
				MaximumPageSize: 20,
				NextPageToken:   pageToken,
				StartTimeFilter: &filterpb.StartTimeFilter{
					LatestTime: timestamp.TimePtr(time.Now().UTC()),
				},
			})
	}
}

type commandFn func(ctx context.Context, pageToken []byte) (Responser, error)

func printCommand(ctx context.Context, fn commandFn, statusCode string) error {
	var pageToken []byte
	var hasMorePage = true
	for hasMorePage {
		resp, err := fn(ctx, pageToken)
		if err != nil {
			return fmt.Errorf("list Temporal workflows: %w", err)
		}
		pageToken = resp.GetNextPageToken()
		hasMorePage = len(pageToken) > 0
		printList(resp, statusCode)
	}
	return nil
}

type Responser interface {
	GetExecutions() []*v110.WorkflowExecutionInfo
	GetNextPageToken() []byte
}

func printList(resp Responser, statusCode string) {
	for _, e := range resp.GetExecutions() {
		var meta JobMetadata
		err := converter.GetDefaultDataConverter().FromPayload(
			e.GetMemo().Fields[MetadataMemo], &meta,
		)
		if err != nil {
			meta.ProductID = "N/A"
		}

		printLine([5]string{
			statusCode,
			e.Execution.WorkflowId,
			meta.ProductID,
			strconv.FormatBool(meta.DryRun),
			e.Status.String(),
		})
	}
}

const MetadataMemo = "jobMeta"

func printHeader() {
	printLine([5]string{" ", "JobID", "Product", "DryRun", "Status"})
	fmt.Printf("%s\n", strings.Repeat("-", 104))
}

func printLine(args [5]string) {
	fmt.Printf("%s %-20s\t%-10s\t%-10s\t%-10s\n", args[0], args[1], args[2], args[3], args[4])
}

func getFreeLocalPort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("resolve localhost:0: %w", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("net listen localhost:0: %w", err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

func temporalClient(addr string) (client.Client, error) {
	// disable stdout to remove temporal silly error about log
	stdout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_RDWR, 0644)
	defer func() { os.Stdout = stdout }()

	return client.NewClient(client.Options{
		HostPort: addr,
	})
}

type JobMetadata struct {
	ID        string `json:"job_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	ProductID string `json:"product_id,omitempty"`
	DryRun    bool   `json:"dry_run,omitempty"`
	Estimated int64  `json:"estimated,omitempty"`
}

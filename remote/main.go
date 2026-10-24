// Command remote runs a script on every cka-prep EC2 node using SSM Run Command.
// No SSH, key pair, public IP or open port is needed: the instances only need
// the SSM agent and the AmazonSSMManagedInstanceCore role they already have.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type node struct {
	name string
	id   string
}

func main() {
	region := flag.String("region", "eu-west-3", "AWS region")
	script := flag.String("script", "../install-k8s.sh", "local script to run on every node")
	only := flag.String("nodes", "", "comma-separated node names to target (default: all)")
	timeout := flag.Duration("timeout", 30*time.Minute, "max time to wait for the script on each node")
	flag.Parse()

	if err := run(*region, *script, *only, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(region, scriptPath, only string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	body, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return err
	}

	nodes, err := findNodes(ctx, ec2.NewFromConfig(cfg))
	if err != nil {
		return err
	}
	if only != "" {
		nodes = filterNodes(nodes, only)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no running nodes found in %s", region)
	}

	client := ssm.NewFromConfig(cfg)
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.id
	}

	// The script travels base64-encoded so quotes and heredocs survive intact.
	encoded := base64.StdEncoding.EncodeToString(body)
	sent, err := client.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  ids,
		Comment:      aws.String("cka-prep remote script"),
		Parameters: map[string][]string{
			"commands": {
				"set -eu",
				"export DEBIAN_FRONTEND=noninteractive",
				fmt.Sprintf("echo %s | base64 -d > /tmp/remote-script.sh", encoded),
				// first-boot user data may still be holding the apt lock
				"cloud-init status --wait || true",
				"bash /tmp/remote-script.sh",
			},
			"executionTimeout": {fmt.Sprint(int(timeout.Seconds()))},
		},
	})
	if err != nil {
		return fmt.Errorf("send command (are the instances registered in SSM?): %w", err)
	}
	commandID := aws.ToString(sent.Command.CommandId)
	fmt.Printf("command %s sent to %d nodes\n", commandID, len(nodes))

	results := make([]*ssm.GetCommandInvocationOutput, len(nodes))
	errs := make([]error, len(nodes))
	var wg sync.WaitGroup
	for i, n := range nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = wait(ctx, client, commandID, n)
		}()
	}
	wg.Wait()

	failed := 0
	for i, n := range nodes {
		fmt.Printf("\n===== %s (%s) =====\n", n.name, n.id)
		if errs[i] != nil {
			failed++
			fmt.Println("error:", errs[i])
			continue
		}
		r := results[i]
		fmt.Print(aws.ToString(r.StandardOutputContent))
		if s := aws.ToString(r.StandardErrorContent); s != "" {
			fmt.Println("--- stderr ---")
			fmt.Print(s)
		}
		fmt.Printf("status: %s\n", r.Status)
		if r.Status != ssmtypes.CommandInvocationStatusSuccess {
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d nodes failed", failed, len(nodes))
	}
	fmt.Printf("\ndone: %d nodes OK\n", len(nodes))
	return nil
}

// wait polls until the invocation on one node reaches a final state.
func wait(ctx context.Context, client *ssm.Client, commandID string, n node) (*ssm.GetCommandInvocationOutput, error) {
	var last ssmtypes.CommandInvocationStatus
	for {
		out, err := client.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(n.id),
		})
		var notYet *ssmtypes.InvocationDoesNotExist
		switch {
		case errors.As(err, &notYet):
			// the invocation shows up a moment after SendCommand
		case err != nil:
			return nil, err
		default:
			if out.Status != last {
				fmt.Printf("[%s] %s\n", n.name, out.Status)
				last = out.Status
			}
			switch out.Status {
			case ssmtypes.CommandInvocationStatusPending,
				ssmtypes.CommandInvocationStatusInProgress,
				ssmtypes.CommandInvocationStatusDelayed,
				ssmtypes.CommandInvocationStatusCancelling:
			default:
				return out, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// findNodes looks up the running instances tagged Project=cka-prep (set on the CDK stack).
func findNodes(ctx context.Context, client *ec2.Client) ([]node, error) {
	pager := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("tag:Project"), Values: []string{"cka-prep"}},
			{Name: aws.String("instance-state-name"), Values: []string{"running"}},
		},
	})

	var nodes []node
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range page.Reservations {
			for _, inst := range r.Instances {
				n := node{id: aws.ToString(inst.InstanceId)}
				n.name = n.id
				for _, t := range inst.Tags {
					// CDK sets Name to the construct path, e.g. CdkStack/node-1
					if aws.ToString(t.Key) == "Name" {
						n.name = path.Base(aws.ToString(t.Value))
					}
				}
				nodes = append(nodes, n)
			}
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].name < nodes[j].name })
	return nodes, nil
}

func filterNodes(nodes []node, only string) []node {
	want := map[string]bool{}
	for _, n := range strings.Split(only, ",") {
		want[strings.TrimSpace(n)] = true
	}
	var out []node
	for _, n := range nodes {
		if want[n.name] {
			out = append(out, n)
		}
	}
	return out
}

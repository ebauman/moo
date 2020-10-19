package agent

import (
	"fmt"
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/liggitt/tabwriter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func LoadCommand() *cli.Command{
	return &cli.Command{
		Name: "agent",
		Usage: "options for agents",
		Subcommands: []*cli.Command{
			{
				Name: "list",
				Usage: "list agents",
				Action: listAgents,
			},
		},
	}
}

func listAgents(c *cli.Context) error {
	mooClient, _, err := rpc.SetupClients(c.String("server"), c.Bool("insecure"), c.String("cacerts"))
	if err != nil {
		return err
	}

	var agentStatus rpc.Status
	switch c.Args().Get(0) {
	case "unknown":
		agentStatus = rpc.Status_Unknown
	case "accepted":
		agentStatus = rpc.Status_Accepted
	case "held":
		agentStatus = rpc.Status_Held
	case "denied":
		agentStatus = rpc.Status_Denied
	case "pending":
		agentStatus = rpc.Status_Pending
	case "error":
		agentStatus = rpc.Status_Error
	default:
		log.Fatalf("invalid agent status type %s specified", c.String("status"))
	}

	agents, err := mooClient.ListAgents(c.Context, &rpc.ListRequest{Status: agentStatus})
	if err != nil {
		log.Fatalf("error while calling ListAgents: %s", err)
	}

	printAgents(agents)

	return nil
}

func printAgents(agents *rpc.AgentListResponse) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	defer tabwriter.Flush()

	headers := []string{"ID", "CLUSTER NAME", "SECRET", "USE EXISTING", "IP", "STATUS", "STATUS MESSAGE"}
	_, err := fmt.Fprintf(tabwriter, "%s\n", strings.Join(headers, "\t"))
	if err != nil {
		log.Fatalf("failed to print headers")
	}

	for _, agent := range agents.Agents {
		fmt.Fprintf(tabwriter,"%s\t%s\t%s\t%t\t%s\t%s\t%s\t\n", agent.ID, agent.ClusterName, "[hidden]", agent.UseExisting, agent.IP, agent.Status, agent.StatusMessage)
	}
}
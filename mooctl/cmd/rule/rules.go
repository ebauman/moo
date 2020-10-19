package rule

import (
	"fmt"
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/liggitt/tabwriter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func LoadCommand() *cli.Command {
	return &cli.Command{
		Name: "rule",
		Usage: "options for rules",
		Subcommands: []*cli.Command{
			{
				Name: "list",
				Usage: "list rules",
				Action: listRules,
			},
			{
				Name: "delete",
				Usage: "delete rule",
				Action: deleteRule,
			},
			{
				Name: "create",
				Usage: "create rule",
				Action: createRule,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Usage:    "rule type (all, cluster-name, source-ip, shared-secret)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "action",
						Usage:    "action (accept, hold, deny)",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "priority",
						Usage:    "priority (rules sorted in descending order)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "regex",
						Usage: "regex for rules that accept it (cluster-name, source-ip)",
					},
				},
			},
		},
	}
}

func createRule(c *cli.Context) error {
	_, rulesClient, err := rpc.SetupClients(c.String("server"), c.Bool("insecure"), c.String("cacerts"))
	if err != nil {
		log.Fatal(err)
	}
	var ruleType rpc.RuleType
	var ruleAction rpc.RuleAction

	switch c.String("type") {
	case "all":
		ruleType = rpc.RuleType_All
	case "cluster-name":
		ruleType = rpc.RuleType_ClusterName
	case "shared-secret":
		ruleType = rpc.RuleType_SharedSecret
	case "source-ip":
		ruleType = rpc.RuleType_SourceIP
	default:
		log.Fatalf("invalid rule type %s specified", c.String("type"))
	}

	switch c.String("action") {
	case "accept":
		ruleAction = rpc.RuleAction_Accept
	case "hold":
		ruleAction = rpc.RuleAction_Hold
	case "deny":
		ruleAction = rpc.RuleAction_Deny
	default:
		log.Fatalf("invalid rule action %s specified", c.String("action"))
	}

	rule := &rpc.Rule{
		Type:     ruleType,
		Action:   ruleAction,
		Priority: int32(c.Int("priority")),
		Regex:    c.String("regex"),
	}

	resp, err := rulesClient.AddRule(c.Context, rule)
	if err != nil {
		log.Fatalf("error while calling AddRule: %s", err)
	}

	if resp.Success {
		fmt.Printf("rule created\n")
	} else {
		fmt.Printf("unable to create rule\n")
	}

	return nil
}

func deleteRule(c *cli.Context) error {
	_, rulesClient, err := rpc.SetupClients(c.String("server"), c.Bool("insecure"), c.String("cacerts"))
	if err != nil {
		log.Fatal(err)
	}
	if c.Args().First() == "" {
		return fmt.Errorf("no index specified")
	}

	ri := &rpc.RuleIndex{
		Index: int32(c.Int("index")),
	}

	resp, err := rulesClient.DeleteRule(c.Context, ri)
	if err != nil {
		log.Fatalf("error while calling DeleteRule %s", err)
	}

	if resp.Success {
		fmt.Printf("rule deleted\n")
	} else {
		fmt.Printf("unable to delete rule\n")
	}

	return nil
}

func listRules(c *cli.Context) error {
	_, rulesClient, err := rpc.SetupClients(c.String("server"), c.Bool("insecure"), c.String("cacerts"))
	if err != nil {
		log.Fatal(err)
	}
	rules, err := rulesClient.ListRules(c.Context, &rpc.Empty{})
	if err != nil {
		return err
	}

	printRules(rules)

	return nil
}

func printRules(rules *rpc.RuleList) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	defer tabwriter.Flush()

	headers := []string{"INDEX", "PRIORITY", "TYPE", "ACTION", "REGEX"}
	_, err := fmt.Fprintf(tabwriter, "%s\n", strings.Join(headers, "\t"))
	if err != nil {
		log.Fatalf("failed to print headers")
	}

	for i, rule := range rules.Rules {
		fmt.Fprintf(tabwriter, "%d\t%d\t%s\t%s\t%s\n", i, rule.Priority, rule.Type, rule.Action, rule.Regex)
	}
}
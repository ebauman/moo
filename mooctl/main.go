package main

import (
	"fmt"
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/liggitt/tabwriter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"os"
	"strings"
)

var (
	logger *log.Logger
	mooClient rpc.MooClient
	rulesClient rpc.RulesClient
)

func init() {
	logger = log.New()
}

func main() {
	app := &cli.App {
		Name: "mooctl",
		Usage: "manage moo servers",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "server",
				Usage: "moo server hostname",
				EnvVars: []string{"MOO_SERVER"},
			},
			&cli.BoolFlag{
				Name: "insecure",
				Usage: "insecure connection to moo server",
				EnvVars: []string{"MOO_SERVER_INSECURE"},
			},
		},
		Commands: []*cli.Command{
			{
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
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name: "index",
								Usage: "index of the rule to delete",
								Required: true,
							},
						},
					},
					{
						Name: "create",
						Usage: "create rule",
						Action: createRule,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "type",
								Usage: "rule type (all, cluster-name, source-ip, shared-secret)",
								Required: true,
							},
							&cli.StringFlag{
								Name: "action",
								Usage: "action (accept, hold, deny)",
								Required: true,
							},
							&cli.IntFlag{
								Name: "priority",
								Usage: "priority (rules sorted in descending order)",
								Required: true,
							},
							&cli.StringFlag{
								Name: "regex",
								Usage: "regex for rules that accept it (cluster-name, source-ip)",
							},
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

func setupClient(hostname string, insecure bool) {
	var conn *grpc.ClientConn
	var err error

	if insecure {
		conn, err = grpc.Dial(hostname, grpc.WithInsecure())
	} else {
		conn, err = grpc.Dial(hostname)
	}

	if err != nil {
		logger.Fatalf("error building moo client: %s", err)
	}

	mooClient = rpc.NewMooClient(conn)
	rulesClient = rpc.NewRulesClient(conn)
}

func listRules(c *cli.Context) error {
	setupClient(c.String("server"), c.Bool("insecure"))

	rules, err := rulesClient.ListRules(c.Context, &rpc.Empty{})
	if err != nil {
		return err
	}

	printRules(rules)

	return nil
}

func deleteRule(c *cli.Context) error {
	setupClient(c.String("server"), c.Bool("insecure"))

	ri := &rpc.RuleIndex{
		Index: int32(c.Int("index")),
	}

	resp, err := rulesClient.DeleteRule(c.Context, ri)
	if err != nil {
		logger.Fatalf("error while calling DeleteRule %s", err)
	}

	if resp.Success {
		fmt.Printf("rule deleted\n")
	} else {
		fmt.Printf("unable to delete rule\n")
	}

	return nil
}

func createRule(c *cli.Context) error {
	setupClient(c.String("server"), c.Bool("insecure"))

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
		logger.Fatalf("invalid rule type %s specified", c.String("type"))
	}

	switch c.String("action") {
	case "accept":
		ruleAction = rpc.RuleAction_Accept
	case "hold":
		ruleAction = rpc.RuleAction_Hold
	case "deny":
		ruleAction = rpc.RuleAction_Deny
	default:
		logger.Fatalf("invalid rule action %s specified", c.String("action"))
	}

	rule := &rpc.Rule{
		Type:     ruleType,
		Action:   ruleAction,
		Priority: int32(c.Int("priority")),
		Regex:    c.String("regex"),
	}

	resp, err := rulesClient.AddRule(c.Context, rule)
	if err != nil {
		logger.Fatalf("error while calling AddRule: %s", err)
	}

	if resp.Success {
		fmt.Printf("rule created\n")
	} else {
		fmt.Printf("unable to create rule\n")
	}

	return nil
}

func printRules(rules *rpc.RuleList) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	defer tabwriter.Flush()

	headers := []string{"INDEX", "PRIORITY", "TYPE", "ACTION", "REGEX"}
	_, err := fmt.Fprintf(tabwriter, "%s\n", strings.Join(headers, "\t"))
	if err != nil {
		logger.Fatalf("failed to print headers")
	}

	for i, rule := range rules.Rules {
		fmt.Fprintf(tabwriter, "%d\t%d\t%s\t%s\t%s\n", i, rule.Priority, rule.Type, rule.Action, rule.Regex)
	}
}


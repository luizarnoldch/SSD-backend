package lambdas

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	CONNECT_ID   = os.Getenv("CONNECT_ID")
	METRIC_TABLE_NAME = os.Getenv("METRIC_TABLE_NAME")
)

type Agent struct {
	ID       string
	ARN      string
	Username string
}

func NewAgent(id, arn, username string) *Agent {
	return &Agent{
		ID:       id,
		ARN:      arn,
		Username: username,
	}
}

func connectListAgent(ctx context.Context, client *connect.Client) []*Agent {
	var allAgents []*Agent
    isTokenAvailable := true
    nextToken := ""
    for isTokenAvailable {
        agentResponse, err := client.ListUsers(
            ctx,
            &connect.ListUsersInput{
                InstanceId: aws.String(CONNECT_ID),
                NextToken:  aws.String(nextToken),
            },
        )
        if err != nil {
            log.Printf("Error associating phone number with contact flow: %v", err)
        }
		for _ , agent := range agentResponse.UserSummaryList {
			if agent.Username == nil {
				agentObj := NewAgent(*agent.Id, "", "")
				allAgents = append(allAgents, agentObj)
				continue
			} 
			agentObj := NewAgent(*agent.Id, *agent.Arn, *agent.Username)
			allAgents = append(allAgents, agentObj)
		}
        nextToken = *agentResponse.NextToken
        isTokenAvailable = nextToken != ""
    }
	fmt.Println("TOTAL NUMBER OF AGENTS:", len(allAgents))
    return allAgents
}

func HandleRequest(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	connectClient := connect.NewFromConfig(cfg)
	ddbClient := dynamodb.NewFromConfig(cfg)
	agents := connectListAgent(ctx, connectClient)
	for _ , agent := range agents {
		fmt.Println(agent.ID)
		fmt.Println("Looking for agent:", agent.ARN)
		ddbResp, err := ddbClient.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(METRIC_TABLE_NAME),
			Key: map[string]types.AttributeValue{
				"Identifier": &types.AttributeValueMemberS{Value: agent.ARN},
				"Type":       &types.AttributeValueMemberS{Value: "Agent-Metric"},
			},
		})
		if err != nil {
			fmt.Println("Error fetching item from DynamoDB:", err)
			return "Status 500", err
		}
		fmt.Println("DDB response:", ddbResp)
	}
	return "SUCCESS", nil
}

func main() {
	log.Println("Lambda function started.")
	lambda.Start(HandleRequest)
}
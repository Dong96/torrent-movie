package azureci

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2019-12-01/containerinstance"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	subscriptionID     string
	resourceGroupName  string
	containerGroupName string
)

func init() {
	subscriptionID = "463b455a-01b9-4bf3-8119-3c34a063dca8"
	resourceGroupName = "torrent-encode_group"
	containerGroupName = "encode-ci"
}

func StartEncodeService() (err error) {

	ctx := context.Background()
	resp, err := startContainerGroup(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err != nil {
		return
	}
	log.Println("Start encode service successful!")

	return
}

func StopEncodeService() (err error) {
	ctx := context.Background()
	resp, err := stopContainerGroup(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

func GetStateOfService() (err error, state string) {
	ctx := context.Background()
	listPage, err := listContainerInstance(ctx, resourceGroupName)
	if err != nil {
		return
	}
	for listPage.Values() != nil {
		for _, v := range listPage.Values() {
			if *v.Name == containerGroupName {
				return nil, *v.ProvisioningState
			}
		}
		if err = listPage.NextWithContext(context.Background()); err != nil {
			return
		}
	}
	return
}

func getContainerGroupsClient() (containerinstance.ContainerGroupsClient, error) {
	containerGroupsClient := containerinstance.NewContainerGroupsClient(subscriptionID)
	auth, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("Failed to get OAuth config: %v", err)
	}
	containerGroupsClient.Authorizer = auth
	return containerGroupsClient, nil
}

func createContainerGroup(ctx context.Context, containerGroupName, location, resourceGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}
	future, err := containerGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		containerGroupName,
		containerinstance.ContainerGroup{
			Name:     &containerGroupName,
			Location: &location,
			ContainerGroupProperties: &containerinstance.ContainerGroupProperties{
				IPAddress: &containerinstance.IPAddress{
					Type: containerinstance.Public,
					Ports: &[]containerinstance.Port{
						{
							Port:     to.Int32Ptr(80),
							Protocol: containerinstance.TCP,
						},
					},
				},
				OsType: containerinstance.Linux,
				Containers: &[]containerinstance.Container{
					{
						Name: to.StringPtr("encode-container"),
						ContainerProperties: &containerinstance.ContainerProperties{
							Ports: &[]containerinstance.ContainerPort{
								{
									Port: to.Int32Ptr(80),
								},
							},
							Image: to.StringPtr("nddong1996/encode-video:latest"),
							Resources: &containerinstance.ResourceRequirements{
								Limits: &containerinstance.ResourceLimits{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1),
								},
								Requests: &containerinstance.ResourceRequests{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1),
								},
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return containerinstance.ContainerGroup{}, err
	}

	err = future.WaitForCompletionRef(ctx, containerGroupsClient.Client)
	if err != nil {
		return containerinstance.ContainerGroup{}, err
	}
	return future.Result(containerGroupsClient)
}

func startContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (r autorest.Response, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return r, fmt.Errorf("cannot get container group client: %v", err)
	}

	future, err := containerGroupsClient.Start(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return r, fmt.Errorf("cannot start container group: %v", err)
	}
	return future.Result(containerGroupsClient)
}

func stopContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (r autorest.Response, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return r, fmt.Errorf("cannot get container group client: %v", err)
	}

	resp, err := containerGroupsClient.Stop(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return r, fmt.Errorf("cannot stop container group: %v", err)
	}
	return resp, err
}

func deleteContainerGroup(ctx context.Context, resourceGroupName, containerGroupName string) (c containerinstance.ContainerGroup, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return c, fmt.Errorf("cannot get container group client: %v", err)
	}

	future, err := containerGroupsClient.Delete(ctx, resourceGroupName, containerGroupName)
	if err != nil {
		return
	}
	return future.Result(containerGroupsClient)
}

func listContainerInstance(ctx context.Context, resourceGroupName string) (result containerinstance.ContainerGroupListResultPage, err error) {
	containerGroupsClient, err := getContainerGroupsClient()
	if err != nil {
		return result, fmt.Errorf("cannot get container group client: %v", err)
	}

	result, err = containerGroupsClient.ListByResourceGroup(ctx, resourceGroupName)
	if err != nil {
		return
	}
	return
}

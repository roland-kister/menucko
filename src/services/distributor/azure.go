package distributor

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

const azureDistributorLogPrefix = "[Azure Distributor]"

type AzureDistributor struct {
	BlobConnStr   string
	ContainerName string
	BlobName      string
}

func (d AzureDistributor) Distribute(content *bytes.Buffer) error {
	d.log("Creation blob client using the connection string")
	client, err := azblob.NewClientFromConnectionString(d.BlobConnStr, &azblob.ClientOptions{})
	if err != nil {
		d.err(err)
		return err
	}

	d.log("Checking if container \"%s\" exists", d.ContainerName)
	containerExists, err := d.containerExists(client)
	if err != nil {
		d.err(err)
		return err
	}

	if !containerExists {
		d.log("Creating container \"%s\"", d.ContainerName)
		_, err := client.CreateContainer(context.Background(), d.ContainerName, &azblob.CreateContainerOptions{})
		if err != nil {
			d.err(err)
			return err
		}
	}

	d.log("Uploading content")
	contentType := "text/html"

	_, err = client.UploadStream(context.Background(), d.ContainerName, d.BlobName, content, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	})

	if err != nil {
		d.err(err)
		return err
	}

	return nil
}

func (d AzureDistributor) containerExists(client *azblob.Client) (bool, error) {
	pager := client.NewListContainersPager(&azblob.ListContainersOptions{
		Prefix: &d.ContainerName,
	})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return false, err
		}

		for _, containerItem := range resp.ContainerItems {
			if *containerItem.Name == d.ContainerName {
				return true, nil
			}
		}
	}

	return false, nil
}

func (AzureDistributor) log(format string, v ...any) {
	message := azureDistributorLogPrefix + " " + fmt.Sprintf(format, v...)

	log.Println(message)
}

func (AzureDistributor) err(err error) {
	message := azureDistributorLogPrefix + fmt.Sprintf(" Err: %v", err)

	log.Println(message)
}

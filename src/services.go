package main

import (
	"fmt"
	"menucko/services/dateresolver"
	"menucko/services/distributor"
	"menucko/services/renderer"
	"os"
	"strconv"
)

const weekdayEnv = "MENUCKO_WEEKDAY"
const htmlTemplateEnv = "MENUCKO_HTML_TEMPLATE"
const stylesPathEnv = "MENUCKO_STYLES_PATH"
const commitHashEnv = "MENUCKO_COMMIT_HASH"
const blobConnStrEnv = "MENUCKO_BLOB_CONN_STR"
const blobContNameEnv = "MENUCKO_BLOB_CONT_NAME"
const blobNameEnv = "MENUCKO_BLOB_NAME"

func getDateResolver() (dateresolver.DateResolver, error) {
	staticWeekdayStr := os.Getenv(weekdayEnv)

	if len(staticWeekdayStr) == 0 {
		return dateresolver.ProdDateResolver{}, nil
	}

	staticWeekday, err := strconv.Atoi(staticWeekdayStr)
	if err != nil {
		return nil, fmt.Errorf("env \"%s\" with value \"%s\" is not a valid number", weekdayEnv, staticWeekdayStr)
	}

	return dateresolver.DevDateResolver{WeekdayVal: staticWeekday}, nil
}

func getRenderer() (renderer.Renderer, error) {
	htmlTemplate := os.Getenv(htmlTemplateEnv)
	if len(htmlTemplate) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", htmlTemplateEnv)
	}

	stylesPath := os.Getenv(stylesPathEnv)
	if len(stylesPath) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", stylesPathEnv)
	}

	commitHash := os.Getenv(commitHashEnv)
	if len(stylesPath) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", commitHashEnv)
	}

	dateResolver, err := getDateResolver()
	if err != nil {
		return nil, err
	}

	return renderer.HTMLRenderer{
		DateResolver:     dateResolver,
		TemplateFilePath: htmlTemplate,
		StylesPath:       stylesPath,
		CommitHash:       commitHash,
	}, nil
}

func getDistributor() (distributor.Distributor, error) {
	blobConnStr := os.Getenv(blobConnStrEnv)
	if len(blobConnStr) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", blobConnStrEnv)
	}

	blobContName := os.Getenv(blobContNameEnv)
	if len(blobContName) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", blobContNameEnv)
	}

	blobName := os.Getenv(blobNameEnv)
	if len(blobName) == 0 {
		return nil, fmt.Errorf("env \"%s\" is empty", blobNameEnv)
	}

	if blobConnStr == "local" {
		return distributor.LocalDistributor{
			Directory: blobContName,
			FileName:  blobName,
		}, nil
	}

	return distributor.AzureDistributor{
		BlobConnStr:   blobConnStr,
		ContainerName: blobContName,
		BlobName:      blobName,
	}, nil
}

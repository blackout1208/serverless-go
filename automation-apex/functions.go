package main

import "encoding/json"

type FunctionList struct {
	Functions []Function
}

type Function struct {
	FunctionName string
}

// Get the list of functions
func NewFunctionList() (*FunctionList, error) {
	//in case of having long list of functions, consider using pagination and calling this multiple times
	// check: https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-pagination.html
	data, err := run("aws", "lambda", "list-functions")
	if err != nil {
		return nil, err
	}

	var res FunctionList
	err = json.Unmarshal(data, &res)
	return &res, err
}

// Verify if function is existing in aws list
func (fl *FunctionList) HasFunction(fname string) bool {
	for _, v := range fl.Functions {
		if v.FunctionName == fname {
			return true
		}
	}
	return false
}

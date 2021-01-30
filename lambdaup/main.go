package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

//Type for project configuration
type LProject struct {
	Name   string
	Bucket string
	Role   string //assuming that all lambdas in this project has the same role. If not, map of roles can be created
	path   string //location of the project file
}

// To create a new project configuration
func NewLProject(fname string) (*LProject, error) {
	// Read json file
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Println(201)
		return nil, err
	}

	var res LProject
	if err = json.Unmarshal(data, &res); err != nil {
		log.Println(3012, string(data))
		return nil, err
	}

	// the directory where the json file is, is the root folder of the project
	res.path = path.Dir(fname)
	if strings.HasPrefix(res.Role, "arn:") {
		log.Println(2012)
		//if reading from json file already has role
		return &res, nil
	}
	// get the role map
	rmp, err := RoleMap()
	if err != nil {
		log.Println(2013)
		return nil, err
	}

	// extract the role map
	nRole, ok := rmp[res.Role]
	if !ok {
		log.Println(2014)
		return &res, errors.New("Role not found: " + res.Role)
	}
	res.Role = nRole

	return &res, nil
}

//Runs and uploads the lambda
func (lp *LProject) UploadLambda(name string) error {
	// set full path at fpath = lp.path+name
	fpath := path.Join(lp.path, name)

	// set env vars for go build
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")

	// build function
	fmt.Println("Building : " + fpath + ".go")
	if _, err := run("go", "build", "-o", fpath, fpath+".go"); err != nil {
		log.Println(2020, fpath)
		return err
	}

	// zip the file
	// -j means, ignore the folder path in the name and set the zip file only with the name of the original file
	fmt.Println("Zipping to " + fpath + ".zip")
	if _, err := run("zip", "-j", fpath+".zip", fpath); err != nil {
		log.Println(2021)
		return err
	}

	// prefix the lambda name with the project name
	lamname := lp.Name + "_" + name

	// s3 upload method
	// this upload command uses a progress report while upload. So, we use Command instead of Run
	upcmd := exec.Command("aws", "s3", "cp", fpath+".zip", "s3://"+lp.Bucket+"/"+lamname+".zip")

	// read stdout - read the info "answered" byt the s3 upload
	upOut, err := upcmd.StdoutPipe()
	if err != nil {
		log.Println(2022)
		return err
	}

	// here we start the execution of s3 command
	fmt.Println("Starting Upload of " + lamname)
	if err := upcmd.Start(); err != nil {
		log.Println(2023)
		return err
	}

	// Copy the response from the s3 command to display in the terminal
	io.Copy(os.Stdout, upOut)
	// wait for the copy command to finish
	if err := upcmd.Wait(); err != nil {
		log.Println(2024)
		return err
	}

	fl, err := NewFunctionList()
	if err != nil {
		log.Println(2025)
		return err
	}

	ok := fl.HasFunction(lamname)
	if ok {
		fmt.Println("Updating Function " + lamname)
		resp, err := run("aws", "lambda", "update-function-code", "--function-name", lamname, "--s3-bucket", lp.Bucket, "--s3-key", lamname+".zip")
		if err != nil {
			log.Println(2026)
			return err
		}

		fmt.Println(string(resp))
		return nil
	}

	// create lambda function
	fmt.Println("Creating Function")
	resp, err := run("aws", "lambda", "create-function", "--function-name", lamname, "--runtime", "go1.x", "--role", lp.Role, "--handler", name, "--code", "S3Bucket="+lp.Bucket+",S3Key="+lamname+".zip")
	if err != nil {
		return err
	}

	fmt.Println(string(resp))
	return nil
}

func main() {
	lname := flag.String("n", "", "Name of Lambda")
	confloc := flag.String("c", "project.json", "Locations of Config file")
	flag.Parse()

	proj, err := NewLProject(*confloc)
	if err != nil {
		log.Fatal(11, err)
	}

	err = proj.UploadLambda(*lname)
	if err != nil {
		log.Fatal(22, err)
	}
}

/*
 * Program to monitor a folder and uploads files from that folder to an HTTP server.
 * 
 * This program, which uses curl to upload files, reads a conf file where the params for the 
 * program to operate must exist. That conf file (picservice.conf) must exist in the user's 
 * home folder or be passed as an argument in the command line.
 * 
 * The "dest location" can be any HTTP server. Below "dest_location" is a NextCloud server
 * where a folder called "mypics" has been previously created.
 * 
 * Here is the list of params that must exist in picservice.conf file (with example values):
 * 
 * interval_secs = 60
 * program_to_invoke = /usr/bin/curl
 * folder_to_monitor = /home/peter/Pictures/trippics
 * dest_location = https://cloud.blueriversys.com/remote.php/webdav/mypics
 * username = peter
 * password = your-password-goes-here
 *  
 */ 
package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
    "strconv"
    "os/exec"
    "strings"
    "bytes"
    "bufio"
)


var (
    folder string
    destLocation string
    secs int
    username string
    password string
    program string 
    ticker * time.Ticker
)

type AppConfigProperties map[string]string

func ReadConfigFile(filename string) (AppConfigProperties, error) {
    config := AppConfigProperties{}

    if len(filename) == 0 {
        return config, nil
    }
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if equal := strings.Index(line, "="); equal >= 0 {
            if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
                value := ""
                if len(line) > equal {
                    value = strings.TrimSpace(line[equal+1:])
                }
                config[key] = value
            }
        }
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
        return nil, err
    }

    return config, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func startLoop(done chan bool) {
    // process the first batch
    sendFiles()
    
    // start timer
    ticker = time.NewTicker(time.Duration(secs) * time.Second)

    for {
		select {
			case <-ticker.C:
			  sendFiles()
		}
	}
	done <- true
}

func sendFiles() {
    files, err := ioutil.ReadDir(folder)
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
		if ( !file.IsDir() ) {
            sendAndDelete(file.Name(), folder + "/" + file.Name())
        }
    }
}

func sendAndDelete(dstfile string, srcfullpath string) {
	// replace spaces by ampersands
	destfile := strings.Replace(dstfile, " ", "_", -1) // -1 replaces all ocurrencies
    parts := make([]string, 6)
    parts[0] = "-k"
    parts[1] = "-u"
    parts[2] = username +":"+password
    parts[3] = "-T"
    parts[4] = srcfullpath
    parts[5] = destLocation + "/" + destfile
    
    cmd := exec.Command(program, parts...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr    
    err := cmd.Run()
    
    if err != nil {
        fmt.Printf("Error during the upload process: %s\n", err)
        return
    } 
    
    outStr := string(stdout.Bytes())
    
    if len(outStr) > 0 {
        fmt.Printf("Error during the upload process:\n", outStr)
        return
	}

    // now we're safe to delete	
	fmt.Printf("uploaded %s\n", srcfullpath)
	
	// now delete
	cmd = exec.Command("rm", srcfullpath)
	err = cmd.Run()
	if err != nil {
	   fmt.Printf("Error deleting file %s: %s\n", srcfullpath, err)
	   return
	}
}

func loadConfParams(confFile string) {
	config, err := ReadConfigFile(confFile)
	check(err)
	
	secs, err = strconv.Atoi(config["interval_secs"])
	if err != nil {
		fmt.Printf("interval_secs property unparseable or missing in the conf file.\n")
		os.Exit(1)
	}
	program = config["program_to_invoke"]
	if program == "" {
		fmt.Printf("program property missing in the conf file.\n")
		os.Exit(1)
	}
	folder = config["folder_to_monitor"]
	if folder == "" {
		fmt.Printf("folder_to_monitor property missing in the conf file.\n")
		os.Exit(1)
	}
	destLocation = config["dest_location"]
	if destLocation == "" {
		fmt.Printf("dest_location property missing in the conf file.\n")
		os.Exit(1)
	}
	username = config["username"]
	if username == "" {
		fmt.Printf("username property missing in the conf file.\n")
		os.Exit(1)
	}
	password = config["password"]
	if password == "" {
		fmt.Printf("password_secs property missing in the conf file.\n")
		os.Exit(1)
	}
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func main() {
    if len(os.Args) == 1 {
        home, _ := os.UserHomeDir()
        fileName := home+"/picservice.conf"
        if !fileExists(fileName) {
			fmt.Printf("picservice.conf file must exist in %s or be passed as an argument.\n", home)
			os.Exit(1)
		}
		loadConfParams(fileName)
	} else if !fileExists(os.Args[1]) {
        fmt.Printf("Config file %s not found.\n", os.Args[1])
        os.Exit(1)
	} else {
		fmt.Printf("Using %s config file.\n",os.Args[1])
        loadConfParams(os.Args[1])
    }
	
	fmt.Printf("Folder to monitor: %s\n", folder)
	fmt.Printf("Destination location: %s\n", destLocation)
	fmt.Printf("Interval in secs: %d\n", secs)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Program to invoke: %s\n", program)

	// start service loop
	done := make(chan bool, 1)
	go startLoop(done)
	
	// should never get here
	<- done
}

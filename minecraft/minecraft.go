package minecraft

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/parithon/minecraftd/discord"
)

var (
	server      *exec.Cmd
	serverStdin io.WriteCloser
)

const minecraftnet = "https://www.minecraft.net/en-us/download/server/bedrock"
const downloadRegexStr = `https://minecraft.azureedge.net/bin-linux/[^"]*`
const versionRegexStr = `bedrock-server-(.+).zip`

func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func downloadBedrockServer() (verison *string, err error) {
	log.Println("Gathering latest minecraft version")
	resp, err := http.Get(minecraftnet)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dlregx := regexp.MustCompile(downloadRegexStr)
	verregx := regexp.MustCompile(versionRegexStr)
	downloadUrl := dlregx.FindString(string(body))
	version := verregx.FindStringSubmatch(downloadUrl)[1]

	log.Printf("Version: %s\n", version)

	fileUrl, err := url.Parse(downloadUrl)
	if err != nil {
		return nil, err
	}

	path := fileUrl.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	log.Printf("Downloading latest Minecraft Bedrock version: %s\n", version)
	resp, err = client.Get(downloadUrl)
	if err != nil {
		os.Remove(fileName)
		return nil, err
	}

	defer resp.Body.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(fileName)
		return nil, err
	}

	defer file.Close()

	log.Printf("Unzipping latest Minecraft Bedrock version: %s\n", version)
	if _, err := unzip(fileName, fmt.Sprintf("bedrock-server-%s", version)); err != nil {
		os.Remove(fileName)
		return nil, err
	}

	if err := os.Remove(fileName); err != nil {
		log.Println("Failed to remove zip binaries")
	}

	log.Printf("Completed downloading latest Minecraft Bedrock server version: %s", version)

	if server != nil {
		for i := 6; i > 0; i-- {
			serverStdin.Write([]byte(fmt.Sprintf("say shutting down in %d seconds...\n", (i * 5))))
			time.Sleep(time.Second * time.Duration(5))
		}
		stop()
	}

	if _, err := os.Stat("bedrock-server"); err == nil {
		os.Remove("bedrock-server")
	}

	mcdir := fmt.Sprintf("bedrock-server-%s", version)
	if err := os.Symlink(mcdir, "bedrock-server"); err != nil {
		log.Fatal("An error occurred while creating symlink to bedrock-server", err)
	}

	os.WriteFile("bedrock-server/version", []byte(version), 0744)

	return &version, nil
}

func copy(src, dst string) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
	}
	return nil
}

func symlink(name string) {
	data := fmt.Sprintf("/data/%s", name)
	app := fmt.Sprintf("/app/bedrock-server/%s", name)

	if _, err := os.Stat(data); os.IsNotExist(err) {
		if _, err := os.Stat(app); err != nil {
			if err := copy(app, data); err != nil {
				log.Fatal(err)
			}
		}
	}

	if _, err := os.Stat(app); os.IsNotExist(err) {
	} else {
		if err := os.Remove(app); err != nil {
			log.Fatal(err)
		}
	}

	if err := os.Symlink(data, app); err != nil {
		log.Fatalf("Failed to symlink the %s\n%s", data, err)
	}
}

func start(version string) {
	log.Println("Starting bedrock_server...")
	server = exec.Command("./bedrock_server")
	server.Dir = "bedrock-server"
	server.Stdout = log.Writer()

	var err error = nil
	serverStdin, err = server.StdinPipe()
	if err != nil {
		log.Fatal("An error occurred while redirecting Stdin", err)
	}

	if err := server.Start(); err != nil {
		log.Fatal("An error occurred while starting the bedrock_server", err)
	}

	log.Println("Started bedrock_server")

	discord.Started(version)
}

func stop() {
	serverStdin.Write([]byte("say shutting down NOW...\n"))
	time.Sleep(time.Second * time.Duration(5))
	serverStdin.Write([]byte("stop\n"))
	if _, err := server.Process.Wait(); err != nil {
		log.Fatal(err)
	}
	serverStdin.Close()
	server = nil
	serverStdin = nil
	discord.Stopped()
	log.Println("Stopped bedrock_server")
}

func checkForUpdates() (*string, error) {
	log.Println("Checking for Minecraft Bedrock Server updates...")
	versionbytes, err := os.ReadFile("bedrock-server/version")
	if err != nil {
		log.Fatal(err)
	}
	version := string(versionbytes)
	log.Printf("Current: %s", version)

	log.Println("Gathering latest minecraft version")
	resp, err := http.Get(minecraftnet)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dlregx := regexp.MustCompile(downloadRegexStr)
	verregx := regexp.MustCompile(versionRegexStr)
	downloadUrl := dlregx.FindString(string(body))
	onlineversion := verregx.FindStringSubmatch(downloadUrl)[1]
	log.Printf("Available: %s", onlineversion)

	if version != onlineversion {
		log.Println("New version available, downloading...")
		ver, err := downloadBedrockServer()
		if err != nil {
			log.Fatal("An error occurred while downloading the Minecraft Bedrock server", err)
		}

		os.Chmod("bedrock-server/bedrock_server", 0755)

		symlink("worlds")
		symlink("server.properties")
		symlink("permissions.json")
		symlink("whitelist.json")

		start(*ver)
	} else {
		log.Println("No new version available")
	}

	return &version, nil
}

func Startup() error {
	var version *string
	if _, err := os.Stat("bedrock-server"); os.IsNotExist(err) {
		log.Println("Installing latest Minecraft Bedrock Server...")
		version, err = downloadBedrockServer()
		if err != nil {
			log.Fatal("An error occurred while downloading the Minecraft Bedrock server", err)
		}

		os.Chmod("bedrock-server/bedrock_server", 0755)

		symlink("worlds")
		symlink("server.properties")
		symlink("permissions.json")
		symlink("whitelist.json")

	} else {
		version, err = checkForUpdates()
		if err != nil {
			log.Fatal("An error occurred while checking for updates to Minecraft Bedrock server", err)
		}
	}

	start(*version)

	return nil
}

func Shutdown(s os.Signal) error {
	log.Println("Stopping bedrock_server...")

	if s == syscall.SIGQUIT {
		go func() {
			for i := 6; i > 0; i-- {
				serverStdin.Write([]byte(fmt.Sprintf("say shutting down in %d seconds...\n", (i * 5))))
				time.Sleep(time.Second * time.Duration(5))
			}
			stop()
		}()
	} else {
		go func() {
			stop()
		}()
	}

	return nil
}

func Wait() {
	go func() {
		for {
			time.Sleep(time.Hour * time.Duration(12))
			checkForUpdates()
		}
	}()
	for {
		time.Sleep(time.Millisecond * time.Duration(500))
		if server == nil {
			os.Exit(0)
		}
	}
}

func HealthCheck() {
	discord.HealthChecked(true)
}

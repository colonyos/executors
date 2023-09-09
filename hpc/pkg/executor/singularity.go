package executor

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Singularity struct {
	containerDir string
}

type Sif struct {
	DockerImage string
}

func CreateSingularity(containerDir string) *Singularity {
	singularity := &Singularity{containerDir: containerDir}
	return singularity
}

func (singularity *Singularity) Remove(dockerImage string) error {
	return os.Remove(singularity.Sif(dockerImage))
}

func (singularity *Singularity) Build(dockerImage string) (string, error) {
	logMsgs := ""
	sif := &Sif{DockerImage: dockerImage}

	t := template.Must(template.New("sifTemplate").Parse(SifTemplate))
	var defContent bytes.Buffer

	if err := t.Execute(&defContent, sif); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Error executing template")
		return err.Error(), err
	}
	defFile := defContent.String()

	tmpfile, err := ioutil.TempFile("", "singularity_*.def")
	if err != nil {
		return err.Error(), err
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(defFile)
	if err != nil {
		return err.Error(), err
	}
	tmpfile.Close()

	cmdString := "singularity build --fakeroot --force " + singularity.Sif(dockerImage) + " " + tmpfile.Name()
	logMsg := "Building container, running " + cmdString + "\n"
	log.Info(logMsg)
	logMsgs += logMsg

	cmd := exec.Command("bash", "-c", cmdString)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			log.WithFields(log.Fields{"Error": exitError, "Stderr": stderr.String(), "ExitCode": exitError.ExitCode()}).Error("Singularity exited with error")
			logMsg += "Failed to build container image, " + stderr.String()
		} else {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to run singularity command")
			logMsg += "Failed to build container image, " + err.Error()
		}
		return err.Error(), err
	}

	logMsgs += strings.TrimSpace(out.String()) + "\n"
	logMsgs += "Successfully build Singularity container, image save to " + singularity.Sif(dockerImage)

	return logMsgs, nil
}

func (singularity *Singularity) SifExists(dockerImage string) bool {
	sif := singularity.Sif(dockerImage)
	_, err := os.Stat(sif)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (singularity *Singularity) Sif(dockerImage string) string {
	return singularity.containerDir + "/" + dockerImage + ".sif"
}

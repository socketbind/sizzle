package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"os"
	"os/exec"
	"log"
	"time"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)


func fatalIfFailed(err error) {
	if err != nil {
		panic(err)
	}
}

func calculateCompleteUtilization(rootProcess *process.Process) float64 {
	cpuPercent, err := rootProcess.CPUPercent()
	if err != nil {
		return 0
	}

	completeUtilization := cpuPercent

	children, _ := rootProcess.Children()

	if children != nil {
		for _, childProcess := range children {
			completeUtilization += calculateCompleteUtilization(childProcess)
		}
	}

	return completeUtilization
}

func watchProcess(pid int, done chan bool, effect *beep.Streamer, format *beep.Format) {
	watchedProcess, err := process.NewProcess(int32(pid))
	fatalIfFailed(err)

	shouldRun := true

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(*effect)

	for shouldRun {
		utilization := calculateCompleteUtilization(watchedProcess)
		fmt.Printf("Complete utilization: %f\n", utilization)

		select {
			case <- done:
				shouldRun = false
			default:
				time.Sleep(1000 * time.Millisecond)
		}
	}
}

func main() {
	if len(os.Args) > 1 {
		programName := os.Args[1]
		programParams := os.Args[2:]

		cmd := exec.Command(programName, programParams...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		fatalIfFailed(err)

		soundEffect, format, err := wav.Decode(openAsset("data/sizzle.wav"))
		fatalIfFailed(err)

		defer soundEffect.Close()

		done := make(chan bool, 1)

		loopedEffect := InfiniteLoop(soundEffect)

		go watchProcess(cmd.Process.Pid, done, &loopedEffect, &format)

		cmd.Wait()

		done <- true
	} else {
		log.Fatal("Please specify a program")
	}
}
